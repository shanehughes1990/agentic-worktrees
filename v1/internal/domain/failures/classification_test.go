package failures

import (
	"errors"
	"testing"
)

func TestWrapTransientAssignsClass(t *testing.T) {
	err := WrapTransient(errors.New("transient failure"))
	if !IsClass(err, ClassTransient) {
		t.Fatalf("expected transient class, got %q", ClassOf(err))
	}
}

func TestWrapTerminalAssignsClass(t *testing.T) {
	err := WrapTerminal(errors.New("terminal failure"))
	if !IsClass(err, ClassTerminal) {
		t.Fatalf("expected terminal class, got %q", ClassOf(err))
	}
}

func TestWrapPreservesExistingClass(t *testing.T) {
	terminal := WrapTerminal(errors.New("already classified"))
	rewrapped := WrapTransient(terminal)
	if !IsClass(rewrapped, ClassTerminal) {
		t.Fatalf("expected terminal class to be preserved, got %q", ClassOf(rewrapped))
	}
}
