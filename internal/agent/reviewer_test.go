package agent

import (
	"testing"
)

func TestReviewParseSuccess(t *testing.T) {
	jsonStr := `{
		"passed": true,
		"score": 4.5,
		"summary": "所有检查项通过",
		"should_retry": false,
		"details": [
			{"item": "格式检查", "passed": true, "detail": "JSON格式正确"}
		]
	}`
	result, err := parseReviewResponse(jsonStr)
	if err != nil {
		t.Fatalf("parseReviewResponse failed: %v", err)
	}
	if !result.Passed {
		t.Error("expected passed=true")
	}
	if result.Score != 4.5 {
		t.Errorf("expected 4.5, got %f", result.Score)
	}
	if result.ShouldRetry {
		t.Error("expected should_retry=false")
	}
	if len(result.CheckResults) != 1 {
		t.Errorf("expected 1 check result, got %d", len(result.CheckResults))
	}
}

func TestReviewParseFailed(t *testing.T) {
	jsonStr := `{
		"passed": false,
		"score": 2.0,
		"summary": "格式检查未通过",
		"should_retry": true,
		"details": []
	}`
	result, err := parseReviewResponse(jsonStr)
	if err != nil {
		t.Fatalf("parseReviewResponse failed: %v", err)
	}
	if result.Passed {
		t.Error("expected passed=false")
	}
	if result.Score != 2.0 {
		t.Errorf("expected 2.0, got %f", result.Score)
	}
	if !result.ShouldRetry {
		t.Error("expected should_retry=true")
	}
}

func TestReviewParseNoJSON(t *testing.T) {
	_, err := parseReviewResponse("plain text without json")
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
}

func TestBuildReviewPrompt(t *testing.T) {
	r := NewReviewer("test-model")
	checkpoints := []string{"格式检查", "内容检查", "长度检查"}
	prompt := r.buildReviewPrompt(map[string]string{"key": "value"}, checkpoints)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if len(prompt) < 100 {
		t.Errorf("prompt too short: %d chars", len(prompt))
	}
}

func TestReviewerReviewWithResult(t *testing.T) {
	// With mock, the callLLM will return error since no real LLM
	// This tests the error handling path
	r := NewReviewer("test-model")
	result, err := r.Review(nil, map[string]string{"test": "data"}, []string{"check1"})
	if err != nil {
		// Error expected since no LLM - test the result shape
		if result == nil {
			t.Log("expected nil result on error") // callLLM returns error
		}
	}
	if result != nil {
		if result.ShouldRetry != true {
			t.Errorf("expected should_retry=true on error, got %v", result.ShouldRetry)
		}
	}
}

func TestNewReviewer(t *testing.T) {
	r := NewReviewer("gemini-2.0-flash")
	if r == nil {
		t.Fatal("NewReviewer returned nil")
	}
	if r.model != "gemini-2.0-flash" {
		t.Errorf("wrong model: %s", r.model)
	}
}
