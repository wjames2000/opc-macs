package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ModelConfig struct {
	Provider    string  `yaml:"provider"`
	Name        string  `yaml:"name"`
	Temperature float32 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	APIBaseURL  string  `yaml:"api_base_url"`  // 兼容 OpenAI API 格式的地址
	APIKey      string  `yaml:"api_key"`       // API 密钥
}

type RuntimeConfig struct {
	PluginsDir string `yaml:"plugins_dir"`
}

type MemoryConfig struct {
	Engine              string  `yaml:"engine"`
	TopK                int     `yaml:"top_k"`
	SimilarityThreshold float32 `yaml:"similarity_threshold"`
	StorePath           string  `yaml:"store_path"`
	PGConnStr           string  `yaml:"pg_conn_str,omitempty"`
}

type HarnessConfig struct {
	MaxStepsPerAgent int      `yaml:"max_steps_per_agent"`
	HITLOperations   []string `yaml:"hitl_operations"`
	ForbiddenWords   []string `yaml:"forbidden_words"`
}

type Config struct {
	App     AppConfig     `yaml:"app"`
	Model   ModelConfig   `yaml:"model"`
	Runtime RuntimeConfig `yaml:"runtime"`
	Memory  MemoryConfig  `yaml:"memory"`
	Harness HarnessConfig `yaml:"harness"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file %s failed: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: yaml parse failed: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Model.Provider == "" {
		return fmt.Errorf("config: model.provider is required")
	}
	if c.Runtime.PluginsDir == "" {
		c.Runtime.PluginsDir = "./build/plugins"
	}
	if c.Memory.TopK <= 0 {
		c.Memory.TopK = 5
	}
	if c.Memory.SimilarityThreshold <= 0 {
		c.Memory.SimilarityThreshold = 0.7
	}
	if c.Memory.StorePath == "" {
		c.Memory.StorePath = "./data/memory.json"
	}
	if c.Harness.MaxStepsPerAgent <= 0 {
		c.Harness.MaxStepsPerAgent = 10
	}
	if c.Model.Temperature == 0 {
		c.Model.Temperature = 0.3
	}
	if c.Model.MaxTokens == 0 {
		c.Model.MaxTokens = 4096
	}
	return nil
}
