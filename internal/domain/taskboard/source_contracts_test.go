package taskboard

import "testing"

func TestSourceListEntryValidateBasics(t *testing.T) {
	valid := SourceListEntry{
		Identity: SourceIdentity{
			Kind:    SourceKindFile,
			Locator: "docs/overview.md",
		},
		RelativePath: "overview.md",
	}
	if err := valid.ValidateBasics(); err != nil {
		t.Fatalf("expected valid source list entry, got error: %v", err)
	}

	invalid := SourceListEntry{}
	if err := invalid.ValidateBasics(); err == nil {
		t.Fatalf("expected missing source identity validation error")
	}
}
