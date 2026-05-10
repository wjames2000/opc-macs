package hitl

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func testHandler(input, output *bytes.Buffer) *Handler {
	return NewInteractiveHandler(input, output)
}

func TestConfirmYes(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	input.WriteString("y\n")

	h := testHandler(&input, &output)
	result, err := h.Confirm(context.Background(), Operation{Type: "test", Description: "test op"})
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}
	if !result {
		t.Error("expected true for 'y'")
	}
}

func TestConfirmNo(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	input.WriteString("n\n")

	h := testHandler(&input, &output)
	result, err := h.Confirm(context.Background(), Operation{Type: "test", Description: "test op"})
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}
	if result {
		t.Error("expected false for 'n'")
	}
}

func TestConfirmYesFullWord(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	input.WriteString("yes\n")

	h := testHandler(&input, &output)
	result, err := h.Confirm(context.Background(), Operation{Type: "test", Description: "test op"})
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}
	if !result {
		t.Error("expected true for 'yes'")
	}
}

func TestConfirmUnknownInput(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	input.WriteString("maybe\n")

	h := testHandler(&input, &output)
	result, err := h.Confirm(context.Background(), Operation{Type: "test", Description: "test op"})
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}
	if result {
		t.Error("expected false for unknown input")
	}
}

func TestConfirmOutputContainsOperation(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	input.WriteString("y\n")

	h := testHandler(&input, &output)
	h.Confirm(context.Background(), Operation{Type: "send_email", Description: "send test email"})

	outputStr := output.String()
	if !strings.Contains(outputStr, "send_email") {
		t.Error("output should contain operation type")
	}
	if !strings.Contains(outputStr, "send test email") {
		t.Error("output should contain operation description")
	}
}

func TestAutoConfirmNonInteractive(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	// bytes.Buffer is not *os.File → autoConfirm = true
	h := NewHandler(&input, &output)
	result, err := h.Confirm(context.Background(), Operation{Type: "test", Description: "test"})
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}
	if !result {
		t.Error("expected auto-confirm=true for non-interactive reader")
	}
}
