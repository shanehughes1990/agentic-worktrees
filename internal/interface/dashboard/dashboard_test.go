package dashboard

import "testing"

func TestNewUI(t *testing.T) {
	ui := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, ".")
	if ui == nil {
		t.Fatalf("expected ui instance")
	}
	ui.Stop()
}
