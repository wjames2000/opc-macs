.PHONY: all build-main build-plugins test test-all clean run lint vet new-plugin

GO := $(shell which go 2>/dev/null || echo "/usr/local/Cellar/go/1.26.2/bin/go")
BUILD_DIR := build
PLUGINS_DIR := $(BUILD_DIR)/plugins

all: build-main build-plugins

build-main:
	@echo "Building main binary..."
	@mkdir -p $(BUILD_DIR)
	GO111MODULE=on CGO_ENABLED=0 $(GO) build -o $(BUILD_DIR)/opc-agent -ldflags="-s -w" ./cmd/opc-agent/
	@echo "✅ $(BUILD_DIR)/opc-agent"

build-plugins:
	@echo "Building plugins..."
	@mkdir -p $(PLUGINS_DIR)
	cd plugins/copywriter && $(GO) build -buildmode=plugin -o ../../$(PLUGINS_DIR)/copywriter.so . && echo "  ✅ copywriter.so"
	cd plugins/email_sorter && $(GO) build -buildmode=plugin -o ../../$(PLUGINS_DIR)/email_sorter.so . && echo "  ✅ email_sorter.so"
	cd plugins/xhs_poster && $(GO) build -buildmode=plugin -o ../../$(PLUGINS_DIR)/xhs_poster.so . && echo "  ✅ xhs_poster.so"
	@echo "✅ All plugins built"

run: build-main build-plugins
	@echo "Starting OPC-Agent..."
	cd $(BUILD_DIR) && ./opc-agent --config=../config.yaml

test:
	$(GO) test ./internal/... -v -count=1 -short

test-all:
	$(GO) test ./internal/... -v -count=1 -race -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out | tail -5

vet:
	$(GO) vet ./internal/... ./cmd/... ./plugins/...

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./internal/... || echo "golangci-lint not installed, skipping"

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	rm -f data/memory.json data/memory.json.tmp

install-deps:
	$(GO) get google.golang.org/adk@v1.0.0
	$(GO) get github.com/google/uuid
	$(GO) get gopkg.in/yaml.v3
	$(GO) mod tidy

new-plugin:
	@if [ -z "$(NAME)" ]; then \
		echo "用法: make new-plugin NAME=plugin_name SUMMARY='简述' TAGS='tag1 tag2'"; \
		echo "示例: make new-plugin NAME=seo_optimizer SUMMARY='SEO优化' TAGS='seo marketing'"; \
		exit 1; \
	fi
	@scripts/create-plugin.sh "$(NAME)" "$(SUMMARY)" $(TAGS)
