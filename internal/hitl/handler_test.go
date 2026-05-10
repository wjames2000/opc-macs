package hitl

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestConfirmYes(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	input.WriteString("y\n")

	h := NewHandler(&input, &output)
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

	h := NewHandler(&input, &output)
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

	h := NewHandler(&input, &output)
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

	h := NewHandler(&input, &output)
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

	h := NewHandler(&input, &output)
	h.Confirm(context.Background(), Operation{Type: "send_email", Description: "send test email"})

	outputStr := output.String()
	if !strings.Contains(outputStr, "send_email") {
		t.Error("output should contain operation type")
	}
	if !strings.Contains(outputStr, "send test email") {
		t.Error("output should contain operation description")
	}
}
