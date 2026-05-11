package workflow

import (
	"fmt"
	"regexp"
	"strings"
)

var varRe = regexp.MustCompile(`\$\{([^}]+)\}`)

func ResolveVariables(template string, ctx map[string]interface{}) string {
	return varRe.ReplaceAllStringFunc(template, func(match string) string {
		path := match[2 : len(match)-1]
		val, err := resolvePath(ctx, path)
		if err != nil {
			return match
		}
		return fmt.Sprintf("%v", val)
	})
}

func resolvePath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch m := current.(type) {
		case map[string]interface{}:
			if v, ok := m[part]; ok {
				current = v
			} else {
				return nil, fmt.Errorf("key %s not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot index into %T", current)
		}
	}
	return current, nil
}

func evalCondition(condition string, ctx map[string]interface{}) (bool, error) {
	condition = ResolveVariables(condition, ctx)
	// Simple condition evaluation: "value == 'expected'" or "value != 'expected'"
	parts := strings.SplitN(condition, " ", 3)
	if len(parts) != 3 {
		return true, nil // Default: pass through
	}

	left := strings.Trim(parts[0], "'\" ")
	op := parts[1]
	right := strings.Trim(parts[2], "'\" ")

	switch op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	default:
		return true, nil
	}
}
