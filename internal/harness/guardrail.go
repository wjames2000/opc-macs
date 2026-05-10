package harness

import "fmt"

type GuardrailType int

const (
	GuardrailBlockWord   GuardrailType = iota
	GuardrailMaxSteps
	GuardrailSensitiveOp
)

type GuardrailAction int

const (
	ActionBlock      GuardrailAction = iota
	ActionHITLConfirm
	ActionLogWarn
)

func (a GuardrailAction) Priority() int {
	switch a {
	case ActionBlock:
		return 3
	case ActionHITLConfirm:
		return 2
	case ActionLogWarn:
		return 1
	default:
		return 0
	}
}

type GuardrailRule struct {
	Type    GuardrailType
	Pattern string
	Action  GuardrailAction
}

type GuardrailResult struct {
	Triggered bool
	Action    GuardrailAction
	Rule      *GuardrailRule
	Message   string
}

type GuardrailEngine struct {
	rules []GuardrailRule
}

func NewGuardrailEngine(rules []GuardrailRule) *GuardrailEngine {
	return &GuardrailEngine{rules: rules}
}

func (ge *GuardrailEngine) AddRule(rule GuardrailRule) {
	ge.rules = append(ge.rules, rule)
}

func (ge *GuardrailEngine) Check(checkItem interface{}) *GuardrailResult {
	var best *GuardrailResult

	for _, rule := range ge.rules {
		var triggered bool

		switch rule.Type {
		case GuardrailBlockWord:
			if s, ok := checkItem.(string); ok {
				triggered = containsPattern(s, rule.Pattern)
			}
		case GuardrailSensitiveOp:
			if s, ok := checkItem.(string); ok {
				triggered = containsPattern(s, rule.Pattern)
			}
		default:
			triggered = false
		}

		if triggered {
			r := &GuardrailResult{
				Triggered: true,
				Action:    rule.Action,
				Rule:      &rule,
				Message:   fmt.Sprintf("guardrail triggered: %s", rule.Pattern),
			}
			if best == nil || rule.Action.Priority() > best.Action.Priority() {
				best = r
			}
		}
	}

	if best == nil {
		return &GuardrailResult{Triggered: false}
	}
	return best
}

func containsPattern(s, pattern string) bool {
	return len(pattern) > 0 && containsFold(s, pattern)
}

func containsFold(s, substr string) bool {
	sLower := toLower(s)
	subLower := toLower(substr)
	return len(subLower) > 0 && containsStr(sLower, subLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
