package ingestion

import "testing"

func TestSourceIdentityValidateBasics(t *testing.T) {
	if err := (SourceIdentity{Kind: SourceKindFile, Locator: "docs/scope.md"}).ValidateBasics(); err != nil {
		t.Fatalf("expected valid source identity, got error: %v", err)
	}
	if err := (SourceIdentity{Locator: "docs/scope.md"}).ValidateBasics(); err == nil {
		t.Fatalf("expected missing source kind validation error")
	}
	if err := (SourceIdentity{Kind: SourceKind("   "), Locator: "docs/scope.md"}).ValidateBasics(); err == nil {
		t.Fatalf("expected blank source kind validation error")
	}
	if err := (SourceIdentity{Kind: SourceKindFolder}).ValidateBasics(); err == nil {
		t.Fatalf("expected missing source locator validation error")
	}
	if err := (SourceIdentity{Kind: SourceKindFolder, Locator: "   "}).ValidateBasics(); err == nil {
		t.Fatalf("expected blank source locator validation error")
	}
}

func TestSourceMetadataValidateBasics(t *testing.T) {
	metadata := SourceMetadata{
		Identity: SourceIdentity{
			Kind:    SourceKindFolder,
			Locator: "docs",
		},
		Attributes: map[string]any{
			"walk_depth": 3,
		},
	}
	if err := metadata.ValidateBasics(); err != nil {
		t.Fatalf("expected valid source metadata, got error: %v", err)
	}
	if err := (SourceMetadata{Identity: SourceIdentity{Kind: SourceKindFolder, Locator: "docs"}}).ValidateBasics(); err != nil {
		t.Fatalf("expected valid source metadata without attributes, got error: %v", err)
	}

	invalid := SourceMetadata{}
	if err := invalid.ValidateBasics(); err == nil {
		t.Fatalf("expected missing source identity validation error")
	}
	if err := (SourceMetadata{Identity: SourceIdentity{Kind: SourceKindFolder, Locator: "   "}}).ValidateBasics(); err == nil {
		t.Fatalf("expected source metadata identity validation error for blank locator")
	}
}
