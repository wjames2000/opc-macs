//go:build integration

package runtime

// This test requires actual .so plugin files.
// Run: go test -tags integration -v ./internal/runtime/
//
// The .so files must exist in ../../build/plugins/ relative to this file.
// Build them first with: make build-plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationLoadAllPlugins(t *testing.T) {
	pluginsDir := filepath.Join("..", "..", "build", "plugins")

	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		t.Skip("plugins directory not found, build plugins first: make build-plugins")
	}

	loader := NewLoader(pluginsDir)
	if err := loader.LoadAll(); err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if loader.Count() == 0 {
		t.Fatal("expected at least 1 plugin to load")
	}

	list := loader.List()
	for _, info := range list {
		t.Logf("  loaded: %s v%s", info.Name, info.Version)
	}

	t.Logf("total: %d plugins loaded", loader.Count())
}

func TestIntegrationGetPlugin(t *testing.T) {
	pluginsDir := filepath.Join("..", "..", "build", "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		t.Skip("plugins directory not found")
	}

	loader := NewLoader(pluginsDir)
	loader.LoadAll()

	// Try to find each expected plugin
	for _, name := range []string{"copywriter", "email_sorter", "xhs_poster"} {
		plugin, ok := loader.Get(name)
		if !ok {
			t.Errorf("expected plugin '%s' to be loaded", name)
			continue
		}
		info := plugin.Info()
		if info.Name != name {
			t.Errorf("expected name %s, got %s", name, info.Name)
		}
		if info.Version == "" {
			t.Errorf("expected non-empty version for %s", name)
		}
	}
}
