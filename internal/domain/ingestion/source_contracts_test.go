package ingestion

import (
	"context"
	"errors"
	"testing"
)

type sourceReaderStub struct {
	read func(ctx context.Context, source SourceIdentity) ([]byte, error)
}

func (stub sourceReaderStub) Read(ctx context.Context, source SourceIdentity) ([]byte, error) {
	return stub.read(ctx, source)
}

func TestSourceListEntryValidateBasics(t *testing.T) {
	valid := SourceListEntry{
		Identity: SourceIdentity{
			Kind:    SourceKindFile,
			Locator: "docs/overview.md",
		},
		RelativePath: "overview.md",
		Metadata: &SourceMetadata{
			Identity: SourceIdentity{
				Kind:    SourceKindFile,
				Locator: "docs/overview.md",
			},
			Attributes: map[string]any{
				"relative_path": "overview.md",
			},
		},
	}
	if err := valid.ValidateBasics(); err != nil {
		t.Fatalf("expected valid source list entry, got error: %v", err)
	}

	invalid := SourceListEntry{}
	if err := invalid.ValidateBasics(); err == nil {
		t.Fatalf("expected missing source identity validation error")
	}
}

func TestSourceReaderReadContractIsProviderAgnostic(t *testing.T) {
	providerSource := SourceIdentity{
		Kind:    SourceKind("github-blob"),
		Locator: "org/repo/docs/overview.md@HEAD",
	}
	expectedContent := []byte("content-from-provider")

	var reader SourceReader = sourceReaderStub{
		read: func(ctx context.Context, source SourceIdentity) ([]byte, error) {
			if source != providerSource {
				t.Fatalf("expected source identity to be passed through unchanged, got %#v", source)
			}
			if ctx == nil {
				t.Fatalf("expected non-nil context")
			}
			return expectedContent, nil
		},
	}

	content, err := reader.Read(context.Background(), providerSource)
	if err != nil {
		t.Fatalf("expected successful content read, got error: %v", err)
	}
	if string(content) != string(expectedContent) {
		t.Fatalf("expected content %q, got %q", expectedContent, content)
	}

	readErr := errors.New("provider read failed")
	reader = sourceReaderStub{
		read: func(_ context.Context, _ SourceIdentity) ([]byte, error) {
			return nil, readErr
		},
	}

	content, err = reader.Read(context.Background(), providerSource)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected read error %v, got %v", readErr, err)
	}
	if content != nil {
		t.Fatalf("expected nil content when read fails, got %q", content)
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

func TestSourceListAndSelectionSemanticsUseIdentityAsStableSelectionKey(t *testing.T) {
	entries := []SourceListEntry{
		{
			Identity: SourceIdentity{
				Kind:    SourceKindFile,
				Locator: "provider://doc-1",
			},
		},
		{
			Identity: SourceIdentity{
				Kind:    SourceKindFile,
				Locator: "provider://doc-2",
			},
			RelativePath: "docs/overview.md",
		},
		{
			Identity: SourceIdentity{
				Kind:    SourceKindFile,
				Locator: "provider://doc-3",
			},
			RelativePath: "docs/overview.md",
		},
	}

	for _, entry := range entries {
		if err := entry.ValidateBasics(); err != nil {
			t.Fatalf("expected listed source entry to satisfy basics, got error: %v", err)
		}
	}

	contentsByLocator := map[string][]byte{
		"provider://doc-1": []byte("first"),
		"provider://doc-2": []byte("second"),
		"provider://doc-3": []byte("third"),
	}
	reader := sourceReaderStub{
		read: func(_ context.Context, source SourceIdentity) ([]byte, error) {
			content, ok := contentsByLocator[source.Locator]
			if !ok {
				return nil, errors.New("unknown source locator")
			}
			return content, nil
		},
	}

	for _, entry := range entries {
		content, err := reader.Read(context.Background(), entry.Identity)
		if err != nil {
			t.Fatalf("expected listed source identity to remain selectable, got error: %v", err)
		}
		expected := string(contentsByLocator[entry.Identity.Locator])
		if string(content) != expected {
			t.Fatalf("expected selection by locator %q to return %q, got %q", entry.Identity.Locator, expected, string(content))
		}
	}
}
