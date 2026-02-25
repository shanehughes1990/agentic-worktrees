package dashboard

import "testing"

func TestNewUI(t *testing.T) {
	ui := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, ".", 3)
	if ui == nil {
		t.Fatalf("expected ui instance")
	}
	ui.Stop()
}
