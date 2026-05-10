package config

import (
	"os"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	content := []byte(`
app:
  name: "opc-agent"
  version: "0.1.0"
model:
  provider: "gemini"
  name: "gemini-2.0-flash"
  temperature: 0.3
  max_tokens: 4096
runtime:
  plugins_dir: "./test_plugins"
memory:
  engine: "embedded"
  top_k: 3
  similarity_threshold: 0.8
  store_path: "./test_memory.json"
harness:
  max_steps_per_agent: 5
  hitl_operations: ["send_email"]
  forbidden_words: ["test"]
`)
	tmpFile, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(content)
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.App.Name != "opc-agent" {
		t.Errorf("expected opc-agent, got %s", cfg.App.Name)
	}
	if cfg.Model.Provider != "gemini" {
		t.Errorf("expected gemini, got %s", cfg.Model.Provider)
	}
	if cfg.Runtime.PluginsDir != "./test_plugins" {
		t.Errorf("expected ./test_plugins, got %s", cfg.Runtime.PluginsDir)
	}
	if cfg.Memory.TopK != 3 {
		t.Errorf("expected 3, got %d", cfg.Memory.TopK)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateDefaults(t *testing.T) {
	content := []byte(`
app:
  name: "test"
model:
  provider: "openai"
memory: {}
harness: {}
`)
	tmpFile, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(content)
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Memory.TopK != 5 {
		t.Errorf("default TopK should be 5, got %d", cfg.Memory.TopK)
	}
	if cfg.Memory.SimilarityThreshold != 0.7 {
		t.Errorf("default threshold should be 0.7, got %f", cfg.Memory.SimilarityThreshold)
	}
	if cfg.Harness.MaxStepsPerAgent != 10 {
		t.Errorf("default max steps should be 10, got %d", cfg.Harness.MaxStepsPerAgent)
	}
}

func TestValidateMissingProvider(t *testing.T) {
	content := []byte(`app: {name:"test"}`)
	tmpFile, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(content)
	tmpFile.Close()

	_, err := Load(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}
