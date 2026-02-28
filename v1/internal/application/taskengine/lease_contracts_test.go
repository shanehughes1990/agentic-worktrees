package taskengine

import (
	"errors"
	"testing"
	"time"
)

func TestLeaseValidateRequiresOwnershipAndWindow(t *testing.T) {
	base := time.Now()
	err := (Lease{OwnerID: "worker-1", Token: "lease-1", AcquiredAt: base, ExpiresAt: base}).Validate()
	if !errors.Is(err, ErrInvalidLeaseContract) {
		t.Fatalf("expected ErrInvalidLeaseContract, got %v", err)
	}
}

func TestLeaseRenewRejectsNonOwnerRenewal(t *testing.T) {
	base := time.Now()
	lease := Lease{
		OwnerID:    "worker-1",
		Token:      "lease-1",
		AcquiredAt: base,
		ExpiresAt:  base.Add(2 * time.Minute),
	}
	_, err := lease.Renew(LeaseRenewalRequest{
		OwnerID:   "worker-2",
		RenewedAt: base.Add(1 * time.Minute),
		ExpiresAt: base.Add(3 * time.Minute),
	})
	if !errors.Is(err, ErrInvalidLeaseRenewal) {
		t.Fatalf("expected ErrInvalidLeaseRenewal, got %v", err)
	}
}

func TestLeaseRenewRejectsExpiredLease(t *testing.T) {
	base := time.Now()
	lease := Lease{
		OwnerID:    "worker-1",
		Token:      "lease-1",
		AcquiredAt: base,
		ExpiresAt:  base.Add(1 * time.Minute),
	}
	_, err := lease.Renew(LeaseRenewalRequest{
		OwnerID:   "worker-1",
		RenewedAt: base.Add(2 * time.Minute),
		ExpiresAt: base.Add(3 * time.Minute),
	})
	if !errors.Is(err, ErrInvalidLeaseRenewal) {
		t.Fatalf("expected ErrInvalidLeaseRenewal, got %v", err)
	}
}

func TestLeaseRenewExtendsLeaseForCurrentOwner(t *testing.T) {
	base := time.Now()
	lease := Lease{
		OwnerID:    "worker-1",
		Token:      "lease-1",
		AcquiredAt: base,
		ExpiresAt:  base.Add(1 * time.Minute),
	}
	renewed, err := lease.Renew(LeaseRenewalRequest{
		OwnerID:   "worker-1",
		RenewedAt: base.Add(30 * time.Second),
		ExpiresAt: base.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if renewed.OwnerID != lease.OwnerID {
		t.Fatalf("expected owner %q, got %q", lease.OwnerID, renewed.OwnerID)
	}
	if !renewed.ExpiresAt.After(lease.ExpiresAt) {
		t.Fatalf("expected lease expiration to be extended")
	}
}
