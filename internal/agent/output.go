package agent

import (
	"fmt"
	"strings"
)

// FormatOutput formats plugin output data into readable Chinese text
func FormatOutput(data interface{}, contentType string) string {
	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%+v", data)
	}
	switch contentType {
	case "copywriter":
		return formatCopywriting(m)
	case "email_sorter":
		return formatEmail(m)
	case "xhs_poster":
		return formatXHS(m)
	default:
		return formatGeneric(m)
	}
}

func formatCopywriting(m map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  📝 短文案\n")
	sb.WriteString(fmt.Sprintf("    %s\n\n", getStr(m, "short_copy")))
	sb.WriteString("  📄 长文案\n")
	sb.WriteString(fmt.Sprintf("    %s\n\n", getStr(m, "long_copy")))
	sb.WriteString("  📱 社交媒体文案\n")
	sb.WriteString(fmt.Sprintf("    %s\n\n", getStr(m, "social_copy")))
	sb.WriteString(fmt.Sprintf("  🎨 风格：%s\n", getStr(m, "style")))
	return sb.String()
}

func formatEmail(m map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n")
	cat := getStr(m, "category")
	urgency := getStr(m, "urgency")

	icon := "📧"
	switch cat {
	case "投诉":
		icon = "⚠️"
	case "合作":
		icon = "🤝"
	case "垃圾":
		icon = "🚫"
	}
	sb.WriteString(fmt.Sprintf("  %s 分类：%s", icon, cat))
	if urgency == "高" {
		sb.WriteString(" 🔴 紧急")
	} else if urgency == "中" {
		sb.WriteString(" 🟡 中等")
	}
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  理由：%s\n\n", getStr(m, "reason")))

	reply := getStr(m, "reply_suggestion")
	if reply != "" {
		sb.WriteString("  💬 回复建议\n")
		sb.WriteString(fmt.Sprintf("    %s\n", reply))
	}
	return sb.String()
}

func formatXHS(m map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  📕 %s\n\n", getStr(m, "title")))
	sb.WriteString(fmt.Sprintf("  %s\n\n", getStr(m, "body")))
	sb.WriteString(fmt.Sprintf("  标签：%s\n", formatList(m, "hashtags")))
	sb.WriteString(fmt.Sprintf("  配图建议：%s\n", formatList(m, "image_suggestions")))
	sb.WriteString(fmt.Sprintf("  风格：%s\n", getStr(m, "style")))
	return sb.String()
}

func formatGeneric(m map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n")
	for k, v := range m {
		switch val := v.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("  %s：%s\n", k, val))
		case []interface{}:
			items := make([]string, len(val))
			for i, item := range val {
				items[i] = fmt.Sprintf("%v", item)
			}
			sb.WriteString(fmt.Sprintf("  %s：%s\n", k, strings.Join(items, ", ")))
		default:
			sb.WriteString(fmt.Sprintf("  %s：%+v\n", k, v))
		}
	}
	return sb.String()
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func formatList(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if list, ok := v.([]interface{}); ok {
			items := make([]string, len(list))
			for i, item := range list {
				items[i] = fmt.Sprintf("%v", item)
			}
			return strings.Join(items, "  ")
		}
	}
	return ""
}
