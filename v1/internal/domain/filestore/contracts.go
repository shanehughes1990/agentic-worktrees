package filestore

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"fmt"
	"strings"
)

type StoreKind string

const (
	StoreKindGoogleCloudStorage StoreKind = "gcs"
	StoreKindLocalDisk          StoreKind = "local"
)

type DeliveryKind string

const (
	DeliveryKindNone      DeliveryKind = "none"
	DeliveryKindGoogleCDN DeliveryKind = "google-cdn"
)

type Profile struct {
	StoreKind    StoreKind
	DeliveryKind DeliveryKind
	Root         string
}

func (profile Profile) Validate() error {
	if strings.TrimSpace(string(profile.StoreKind)) == "" {
		return failures.WrapTerminal(errors.New("store_kind is required"))
	}
	switch profile.StoreKind {
	case StoreKindGoogleCloudStorage, StoreKindLocalDisk:
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported store_kind %q", profile.StoreKind))
	}
	if strings.TrimSpace(string(profile.DeliveryKind)) == "" {
		return failures.WrapTerminal(errors.New("delivery_kind is required"))
	}
	switch profile.DeliveryKind {
	case DeliveryKindNone, DeliveryKindGoogleCDN:
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported delivery_kind %q", profile.DeliveryKind))
	}
	if strings.TrimSpace(profile.Root) == "" {
		return failures.WrapTerminal(errors.New("root is required"))
	}
	return nil
}

func (profile Profile) ValidateCompatibility(requireCDN bool) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	if profile.StoreKind == StoreKindLocalDisk && profile.DeliveryKind != DeliveryKindNone {
		return failures.WrapTerminal(errors.New("local filestore is only compatible with delivery_kind=none"))
	}
	if requireCDN {
		if profile.StoreKind != StoreKindGoogleCloudStorage {
			return failures.WrapTerminal(errors.New("cdn-dependent flows require store_kind=gcs"))
		}
		if profile.DeliveryKind != DeliveryKindGoogleCDN {
			return failures.WrapTerminal(errors.New("cdn-dependent flows require delivery_kind=google-cdn"))
		}
	}
	return nil
}
