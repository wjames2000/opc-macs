package hitl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

type Operation struct {
	Type        string
	Description string
	Payload     interface{}
}

type Handler struct {
	reader      io.Reader
	writer      io.Writer
	autoConfirm bool
}

func NewHandler(reader io.Reader, writer io.Writer) *Handler {
	// Auto-confirm when non-interactive (piped input)
	autoConfirm := !isInteractive(reader)

	return &Handler{
		reader:      reader,
		writer:      writer,
		autoConfirm: autoConfirm,
	}
}

func isInteractive(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func NewInteractiveHandler(reader io.Reader, writer io.Writer) *Handler {
	return &Handler{
		reader:      reader,
		writer:      writer,
		autoConfirm: false,
	}
}

func (h *Handler) Confirm(ctx context.Context, op Operation) (bool, error) {
	if h.autoConfirm {
		fmt.Fprintf(h.writer, "[HITL] 非交互模式，自动确认操作：%s\n", op.Type)
		return true, nil
	}

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
