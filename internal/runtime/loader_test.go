package runtime

import (
	"context"
	"testing"
)

// mockPlugin implements AgentPlugin for unit testing
type mockPlugin struct {
	name string
	info PluginInfo
}

func (m *mockPlugin) Name() string                        { return m.name }
func (m *mockPlugin) Info() PluginInfo                    { return m.info }
func (m *mockPlugin) Execute(_ context.Context, _ string, _ map[string]interface{}) (*ExecutionResult, error) {
	return &ExecutionResult{}, nil
}
func (m *mockPlugin) Review(_ context.Context, _ interface{}) (*ReviewResult, error) {
	return &ReviewResult{Passed: true}, nil
}

func TestLoaderNew(t *testing.T) {
	l := NewLoader("./testdata")
	if l == nil {
		t.Fatal("NewLoader returned nil")
	}
	if l.Count() != 0 {
		t.Errorf("expected 0, got %d", l.Count())
	}
}

func TestLoaderListAndCount(t *testing.T) {
	l := NewLoader("./testdata")
	// Manually register mock plugins
	l.plugins["a"] = &mockPlugin{name: "a", info: PluginInfo{Name: "a", Version: "1.0"}}
	l.plugins["b"] = &mockPlugin{name: "b", info: PluginInfo{Name: "b", Version: "2.0"}}

	if l.Count() != 2 {
		t.Errorf("expected 2, got %d", l.Count())
	}

	list := l.List()
	if len(list) != 2 {
		t.Errorf("expected 2, got %d", len(list))
	}

	names := make(map[string]bool)
	for _, p := range list {
		names[p.Name] = true
	}
	if !names["a"] || !names["b"] {
		t.Errorf("expected both a and b, got %v", names)
	}
}

func TestLoaderGet(t *testing.T) {
	l := NewLoader("./testdata")
	l.plugins["test"] = &mockPlugin{name: "test"}

	p, ok := l.Get("test")
	if !ok {
		t.Fatal("expected to find 'test'")
	}
	if p.Name() != "test" {
		t.Errorf("expected 'test', got %s", p.Name())
	}

	_, ok = l.Get("nonexistent")
	if ok {
		t.Fatal("expected not to find nonexistent")
	}
}

func TestLoaderLoadAllNoDir(t *testing.T) {
	l := NewLoader("/nonexistent/plugins")
	err := l.LoadAll()
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}
