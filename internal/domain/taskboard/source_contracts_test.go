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

func TestSourceListEntryValidateBasicsProviderAgnosticListingSemantics(t *testing.T) {
	entry := SourceListEntry{
		Identity: SourceIdentity{
			Kind:    SourceKindFile,
			Locator: "provider://doc-42",
		},
	}

	if err := entry.ValidateBasics(); err != nil {
		t.Fatalf("expected source entry without relative path to be valid, got error: %v", err)
	}

	entry.RelativePath = "docs/overview.md"
	if err := entry.ValidateBasics(); err != nil {
		t.Fatalf("expected source entry with provider locator and display path to be valid, got error: %v", err)
	}
}
