package taskboard

import (
	"fmt"
	"strings"
)

type SourceKind string

const (
	SourceKindFile   SourceKind = "file"
	SourceKindFolder SourceKind = "folder"
)

type SourceIdentity struct {
	Kind    SourceKind `json:"kind"`
	Locator string     `json:"locator"`
}

type SourceMetadata struct {
	Identity   SourceIdentity `json:"identity"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (identity SourceIdentity) ValidateBasics() error {
	if strings.TrimSpace(string(identity.Kind)) == "" {
		return fmt.Errorf("source kind is required")
	}
	if strings.TrimSpace(identity.Locator) == "" {
		return fmt.Errorf("source locator is required")
	}
	return nil
}

func (metadata SourceMetadata) ValidateBasics() error {
	if err := metadata.Identity.ValidateBasics(); err != nil {
		return err
	}
	return nil
}
