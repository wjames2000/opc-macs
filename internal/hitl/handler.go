package hitl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

type Operation struct {
	Type        string
	Description string
	Payload     interface{}
}

type Handler struct {
	reader io.Reader
	writer io.Writer
}

func NewHandler(reader io.Reader, writer io.Writer) *Handler {
	return &Handler{reader: reader, writer: writer}
}

func (h *Handler) Confirm(ctx context.Context, op Operation) (bool, error) {
	fmt.Fprintf(h.writer, "\n"+
		"═══════════════════════════════════════════════\n"+
		"  ⚠️  需要您的确认\n"+
		"  操作：%s\n"+
		"  详情：%s\n"+
		"═══════════════════════════════════════════════\n"+
		"  请输入 y（同意）/ n（拒绝）：", op.Type, op.Description)

	scanner := bufio.NewScanner(h.reader)
	if !scanner.Scan() {
		return false, fmt.Errorf("hitl: read input failed: %w", scanner.Err())
	}

	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return input == "y" || input == "yes", nil
}
