package agent

import (
	"context"
	"testing"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

// mockLoader for testing Router
type mockLoader struct {
	plugins map[string]runtime.AgentPlugin
}

func newMockLoader() *mockLoader {
	return &mockLoader{plugins: map[string]runtime.AgentPlugin{
		"copywriter":   &testPlugin{name: "copywriter", summary: "文案生成", hitl: true},
		"email_sorter": &testPlugin{name: "email_sorter", summary: "邮件分类", hitl: true},
	}}
}

func (m *mockLoader) Get(name string) (runtime.AgentPlugin, bool) {
	p, ok := m.plugins[name]
	return p, ok
}
func (m *mockLoader) List() []runtime.PluginInfo {
	var result []runtime.PluginInfo
	for _, p := range m.plugins {
		result = append(result, p.Info())
	}
	return result
}
func (m *mockLoader) Count() int { return len(m.plugins) }
func (m *mockLoader) LoadAll() error { return nil }
func (m *mockLoader) loadFile(_ string) error { return nil }

type testPlugin struct {
	name    string
	summary string
	hitl    bool
}

func (t *testPlugin) Name() string { return t.name }
func (t *testPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{Name: t.name, Summary: t.summary, RequiresHITL: t.hitl}
}
func (t *testPlugin) Execute(_ context.Context, _ string, _ map[string]interface{}) (*runtime.ExecutionResult, error) {
	return &runtime.ExecutionResult{}, nil
}
func (t *testPlugin) Review(_ context.Context, _ interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{Passed: true}, nil
}

func TestRouterNew(t *testing.T) {
	loader := newMockLoader()
	r := NewRouter(loader, "gemini-2.0-flash")
	if r == nil {
		t.Fatal("NewRouter returned nil")
	}
}

func TestBuildL1Context(t *testing.T) {
	loader := newMockLoader()
	r := NewRouter(loader, "test-model")
	ctx := r.buildL1Context()
	if ctx == "" {
		t.Fatal("expected non-empty L1 context")
	}
	if len(ctx) < 50 {
		t.Errorf("L1 context too short: %d chars", len(ctx))
	}
}

func TestRouterEmptyInput(t *testing.T) {
	loader := newMockLoader()
	r := NewRouter(loader, "test-model")
	result, err := r.Route(context.Background(), "")
	if err != nil {
		t.Fatalf("Route(empty) should not error: %v", err)
	}
	if result.Action != RouteActionUnknown {
		t.Errorf("expected Unknown for empty input, got %v", result.Action)
	}
	if result.Message == "" {
		t.Error("expected non-empty message for empty input")
	}
}

func TestParseClassifyResponseSuccess(t *testing.T) {
	jsonStr := `{
		"agent": "copywriter",
		"confidence": 0.85,
		"reason": "用户要求写文案"
	}`
	agent, conf, err := parseClassifyResponse(jsonStr)
	if err != nil {
		t.Fatalf("parseClassifyResponse failed: %v", err)
	}
	if agent != "copywriter" {
		t.Errorf("expected copywriter, got %s", agent)
	}
	if conf != 0.85 {
		t.Errorf("expected 0.85, got %f", conf)
	}
}

func TestParseClassifyResponseNoJSON(t *testing.T) {
	_, _, err := parseClassifyResponse("plain text without json")
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
}

func TestBuildAvailableList(t *testing.T) {
	loader := newMockLoader()
	r := NewRouter(loader, "test")
	list := r.buildAvailableList()
	if list == "" {
		t.Fatal("expected non-empty list")
	}
}

func TestRouterUnknownAgent(t *testing.T) {
	loader := newMockLoader()
	r := NewRouter(loader, "test")

	result, err := r.Route(context.Background(), "some random text that wont match")
	if err != nil {
		t.Fatalf("Route should not error on unknown: %v", err)
	}
	if result.Action != RouteActionUnknown {
		t.Errorf("expected Unknown, got %v", result.Action)
	}
}
