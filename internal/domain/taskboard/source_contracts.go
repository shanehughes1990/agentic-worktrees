package taskboard

import "context"

type SourceListOptions struct {
	WalkDepth        int
	IgnorePaths      []string
	IgnoreExtensions []string
}

type SourceListEntry struct {
	Identity     SourceIdentity  `json:"identity"`
	RelativePath string          `json:"relative_path,omitempty"`
	Metadata     *SourceMetadata `json:"metadata,omitempty"`
}

func (entry SourceListEntry) ValidateBasics() error {
	return entry.Identity.ValidateBasics()
}

type SourceLister interface {
	List(ctx context.Context, source SourceMetadata, options SourceListOptions) ([]SourceListEntry, error)
}

type SourceReader interface {
	Read(ctx context.Context, source SourceIdentity) ([]byte, error)
}

type SourceWorkingDirectoryResolver interface {
	ResolveWorkingDirectory(ctx context.Context, source SourceIdentity) (string, error)
}
