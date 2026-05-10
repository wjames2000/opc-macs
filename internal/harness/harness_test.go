package harness

import "testing"

func TestRoleBoundaryValidateAction(t *testing.T) {
	rb := NewRoleBoundary("test", nil, []string{"generate", "review"}, nil, 10)
	if err := rb.ValidateAction(Action{Tool: "generate"}); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if err := rb.ValidateAction(Action{Tool: "delete"}); err == nil {
		t.Fatal("expected error for unauthorized tool")
	}
}

func TestRoleBoundaryValidateOutput(t *testing.T) {
	rb := NewRoleBoundary("test", nil, nil, []string{"secret", "password"}, 10)
	if err := rb.ValidateOutput("hello world"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if err := rb.ValidateOutput("my secret key"); err == nil {
		t.Fatal("expected error for forbidden word")
	}
}

func TestStateMachineValidTransition(t *testing.T) {
	sm := NewStateMachine(10)
	if err := sm.Transition(StateLoadingSkill); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if sm.Current() != StateLoadingSkill {
		t.Errorf("expected loading_skill, got %s", StateName(sm.Current()))
	}
}

func TestStateMachineInvalidTransition(t *testing.T) {
	sm := NewStateMachine(10)
	if err := sm.Transition(StateCompleted); err == nil {
		t.Fatal("expected error for illegal transition from idle to completed")
	}
}

func TestStateMachineMaxSteps(t *testing.T) {
	sm := NewStateMachine(2)
	sm.Transition(StateLoadingSkill)
	sm.Transition(StateRetrievingMemory)
	sm.Transition(StateExecuting) // should fail on 3rd transition
	if sm.Current() != StateFailed {
		t.Errorf("expected failed state, got %s", StateName(sm.Current()))
	}
}

func TestStateMachineListener(t *testing.T) {
	sm := NewStateMachine(10)
	called := false
	sm.Subscribe(func(prev, curr AgentState) {
		called = true
		if prev != StateIdle || curr != StateLoadingSkill {
			t.Errorf("unexpected transition: %s -> %s", StateName(prev), StateName(curr))
		}
	})
	sm.Transition(StateLoadingSkill)
	if !called {
		t.Fatal("listener was not called")
	}
}

func TestArtifactValidation(t *testing.T) {
	valid := &Artifact{SchemaName: "test_schema", Version: "1.0"}
	if err := ValidateArtifact(valid, "test_schema"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if err := ValidateArtifact(valid, "wrong"); err == nil {
		t.Fatal("expected error for wrong schema")
	}
}

func TestGuardrailPriority(t *testing.T) {
	engine := NewGuardrailEngine([]GuardrailRule{
		{Type: GuardrailSensitiveOp, Pattern: "send", Action: ActionHITLConfirm},
		{Type: GuardrailBlockWord, Pattern: "bad", Action: ActionBlock},
	})

	result := engine.Check("please send this bad word")
	if result == nil || !result.Triggered {
		t.Fatal("expected triggered")
	}
	if result.Action != ActionBlock {
		t.Errorf("expected Block (highest priority), got %v", result.Action)
	}
}

func TestGuardrailNoMatch(t *testing.T) {
	engine := NewGuardrailEngine([]GuardrailRule{
		{Type: GuardrailBlockWord, Pattern: "bad", Action: ActionBlock},
	})
	result := engine.Check("clean text")
	if result == nil || result.Triggered {
		t.Fatal("expected not triggered")
	}
}
