package harness

import "fmt"

type AgentState int

const (
	StateIdle             AgentState = iota
	StateLoadingSkill
	StateRetrievingMemory
	StateExecuting
	StateReviewing
	StateHITLCheck
	StateWritingMemory
	StateCompleted
	StateFailed
	StateTimedOut
)

var stateNames = map[AgentState]string{
	StateIdle:             "idle",
	StateLoadingSkill:     "loading_skill",
	StateRetrievingMemory: "retrieving_memory",
	StateExecuting:        "executing",
	StateReviewing:        "reviewing",
	StateHITLCheck:        "hitl_check",
	StateWritingMemory:    "writing_memory",
	StateCompleted:        "completed",
	StateFailed:           "failed",
	StateTimedOut:         "timed_out",
}

var validTransitions = map[AgentState][]AgentState{
	StateIdle:             {StateLoadingSkill},
	StateLoadingSkill:     {StateRetrievingMemory, StateFailed},
	StateRetrievingMemory: {StateExecuting, StateFailed},
	StateExecuting:        {StateReviewing, StateFailed, StateTimedOut},
	StateReviewing:        {StateExecuting, StateHITLCheck, StateFailed},
	StateHITLCheck:        {StateWritingMemory, StateFailed},
	StateWritingMemory:    {StateCompleted, StateFailed},
}

type StateChangeListener func(prev, curr AgentState)

type StateMachine struct {
	current        AgentState
	transitions    int
	maxTransitions int
	listeners      []StateChangeListener
}

func NewStateMachine(maxTransitions int) *StateMachine {
	return &StateMachine{
		current:        StateIdle,
		maxTransitions: maxTransitions,
	}
}

func (sm *StateMachine) Transition(target AgentState) error {
	allowed, ok := validTransitions[sm.current]
	if !ok {
		return fmt.Errorf("harness: no transitions defined from state %s", StateName(sm.current))
	}

	found := false
	for _, s := range allowed {
		if s == target {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("harness: illegal transition from %s to %s",
			StateName(sm.current), StateName(target))
	}

	sm.transitions++
	if sm.transitions > sm.maxTransitions {
		prev := sm.current
		sm.current = StateFailed
		sm.notifyListeners(prev, StateFailed)
		return fmt.Errorf("harness: max transitions exceeded (%d)", sm.maxTransitions)
	}

	prev := sm.current
	sm.current = target
	sm.notifyListeners(prev, target)
	return nil
}

func (sm *StateMachine) Current() AgentState {
	return sm.current
}

func (sm *StateMachine) Subscribe(listener StateChangeListener) {
	sm.listeners = append(sm.listeners, listener)
}

func (sm *StateMachine) notifyListeners(prev, curr AgentState) {
	for _, l := range sm.listeners {
		l(prev, curr)
	}
}

func StateName(s AgentState) string {
	if name, ok := stateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", int(s))
}
