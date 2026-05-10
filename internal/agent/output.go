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
	case "competitive_analysis":
		return formatCompetitive(m)
	case "meeting_minutes":
		return formatMeeting(m)
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

func formatCompetitive(m map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  📊 竞品：%s\n\n", getStr(m, "competitor")))
	sb.WriteString(fmt.Sprintf("  定位：%s\n\n", getStr(m, "market_position")))

	sb.WriteString("  ✅ 优势\n")
	for _, s := range getList(m, "strengths") {
		sb.WriteString(fmt.Sprintf("    • %s\n", s))
	}

	sb.WriteString("\n  ❌ 劣势\n")
	for _, s := range getList(m, "weaknesses") {
		sb.WriteString(fmt.Sprintf("    • %s\n", s))
	}

	sb.WriteString("\n  📈 机会\n")
	for _, s := range getList(m, "opportunities") {
		sb.WriteString(fmt.Sprintf("    • %s\n", s))
	}

	sb.WriteString("\n  ⚠️  威胁\n")
	for _, s := range getList(m, "threats") {
		sb.WriteString(fmt.Sprintf("    • %s\n", s))
	}

	sb.WriteString(fmt.Sprintf("\n  💡 差异化建议\n    %s\n", getStr(m, "differentiation")))
	risk := getStr(m, "risk_level")
	icon := "🟢"
	if risk == "高" {
		icon = "🔴"
	} else if risk == "中" {
		icon = "🟡"
	}
	sb.WriteString(fmt.Sprintf("\n  风险等级：%s %s\n", icon, risk))
	sb.WriteString(fmt.Sprintf("\n  📋 总结\n    %s\n", getStr(m, "summary")))
	return sb.String()
}

func formatMeeting(m map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  📅 %s\n\n", getStr(m, "title")))

	if t := getStr(m, "time"); t != "" {
		sb.WriteString(fmt.Sprintf("  时间：%s\n", t))
	}
	if p := getList(m, "participants"); len(p) > 0 {
		sb.WriteString(fmt.Sprintf("  参会：%s\n\n", strings.Join(p, "、")))
	}

	sb.WriteString("  📋 议程\n")
	for _, a := range getList(m, "agenda") {
		sb.WriteString(fmt.Sprintf("    • %s\n", a))
	}

	sb.WriteString("\n  ✅ 决策\n")
	for _, d := range getList(m, "decisions") {
		sb.WriteString(fmt.Sprintf("    • %s\n", d))
	}

	sb.WriteString("\n  🔧 待办事项\n")
	if items, ok := m["action_items"].([]interface{}); ok {
		for _, item := range items {
			if im, ok := item.(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("    ☐ %s（%s，%s）\n",
					getStr(im, "task"), getStr(im, "owner"), getStr(im, "deadline")))
			}
		}
	}

	if ns := getStr(m, "next_steps"); ns != "" {
		sb.WriteString(fmt.Sprintf("\n  📌 下一步\n    %s\n", ns))
	}
	if kd := getStr(m, "key_discussions"); kd != "" {
		sb.WriteString(fmt.Sprintf("\n  💬 关键讨论\n    %s\n", kd))
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

func getList(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if list, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(list))
			for _, item := range list {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}
