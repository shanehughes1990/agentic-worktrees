package errors

import (
	errorspkg "errors"
	"testing"
)

func TestWrapAndClassOf(t *testing.T) {
	base := errorspkg.New("boom")
	wrapped := Wrap(ClassTransient, "ingest", base)

	if !IsClass(wrapped, ClassTransient) {
		t.Fatalf("expected transient class")
	}
	if ClassOf(wrapped) != ClassTransient {
		t.Fatalf("unexpected class: %s", ClassOf(wrapped))
	}
}

func TestTransientTerminalHelpers(t *testing.T) {
	base := errorspkg.New("boom")
	if ClassOf(Transient("retry", base)) != ClassTransient {
		t.Fatalf("expected transient class")
	}
	if ClassOf(Terminal("fail", base)) != ClassTerminal {
		t.Fatalf("expected terminal class")
	}
}

func TestWrapNilReturnsNil(t *testing.T) {
	if Wrap(ClassTerminal, "x", nil) != nil {
		t.Fatalf("expected nil wrap result for nil error")
	}
}
