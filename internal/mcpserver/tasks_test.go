package mcpserver

import "testing"

func TestTaskTransitions(t *testing.T) {
	if !CanTransitionTask(TaskProposed, TaskReady) {
		t.Fatal("proposed should transition to ready")
	}
	if CanTransitionTask(TaskDone, TaskInProgress) {
		t.Fatal("done must be terminal")
	}
	if CanTransitionTask(TaskReady, TaskDone) {
		t.Fatal("ready should not skip directly to done")
	}
}
