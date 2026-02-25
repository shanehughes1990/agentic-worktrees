package dashboard

import "testing"

func TestNewUI(t *testing.T) {
	ui := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, ".", "revamp", "redis://localhost:6379/0", 3)
	if ui == nil {
		t.Fatalf("expected ui instance")
	}
	ui.Stop()
}
