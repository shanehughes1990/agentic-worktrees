package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationworker "agentic-orchestrator/internal/application/worker"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkerRegistryRegisterHeartbeatStateAndDelete(t *testing.T) {
	registry, err := NewWorkerRegistry(newTestDB(t))
	if err != nil {
		t.Fatalf("new worker registry: %v", err)
	}

	now := time.Now().UTC()
	worker, err := registry.Register(
		context.Background(),
		"worker-1",
		[]taskengine.JobKind{taskengine.JobKindSCMWorkflow},
		now,
		now.Add(30*time.Second),
	)
	if err != nil {
		t.Fatalf("register worker: %v", err)
	}
	if worker.Epoch != 1 {
		t.Fatalf("expected first registration epoch 1, got %d", worker.Epoch)
	}

	err = registry.Upsert(context.Background(), taskengine.WorkerCapabilityAdvertisement{
		WorkerID: "worker-1",
		Capabilities: []taskengine.WorkerCapability{{
			Kind: taskengine.JobKindSCMWorkflow,
		}},
	})
	if err != nil {
		t.Fatalf("upsert worker: %v", err)
	}

	workers, err := registry.ListWorkers(context.Background(), 10)
	if err != nil {
		t.Fatalf("list workers: %v", err)
	}
	if len(workers) != 1 {
		t.Fatalf("expected one worker record after upsert, got %d", len(workers))
	}
	if workers[0].Epoch != 2 {
		t.Fatalf("expected upsert to bump epoch to 2, got %d", workers[0].Epoch)
	}

	heartbeatAt := now.Add(1 * time.Minute)
	touched, err := registry.TouchHeartbeat(context.Background(), "worker-1", 2, heartbeatAt, heartbeatAt.Add(45*time.Second))
	if err != nil {
		t.Fatalf("touch heartbeat: %v", err)
	}
	if touched.LastHeartbeat.Unix() != heartbeatAt.Unix() {
		t.Fatalf("expected heartbeat update to persist, got %s", touched.LastHeartbeat)
	}

	updated, err := registry.UpdateState(context.Background(), "worker-1", 2, domainrealtime.StateInvalidated, heartbeatAt.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("update state: %v", err)
	}
	if updated.State != domainrealtime.StateInvalidated {
		t.Fatalf("expected state invalidated, got %s", updated.State)
	}

	if err := registry.RemoveRegistration(context.Background(), "worker-1", 2); err != nil {
		t.Fatalf("remove registration: %v", err)
	}
	workers, err = registry.ListWorkers(context.Background(), 10)
	if err != nil {
		t.Fatalf("list workers after remove: %v", err)
	}
	if len(workers) != 0 {
		t.Fatalf("expected no workers after remove, got %d", len(workers))
	}
}

func TestWorkerRegistrySettingsUpsertAndRead(t *testing.T) {
	registry, err := NewWorkerRegistry(newTestDB(t))
	if err != nil {
		t.Fatalf("new worker registry: %v", err)
	}

	settings := domainrealtime.Settings{
		HeartbeatInterval: 20 * time.Second,
		ResponseDeadline:  45 * time.Second,
		UpdatedAt:         time.Now().UTC(),
	}

	if _, err := registry.UpsertSettings(context.Background(), settings); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	loaded, err := registry.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if loaded.HeartbeatInterval != settings.HeartbeatInterval {
		t.Fatalf("expected heartbeat interval %s, got %s", settings.HeartbeatInterval, loaded.HeartbeatInterval)
	}
	if loaded.ResponseDeadline != settings.ResponseDeadline {
		t.Fatalf("expected response deadline %s, got %s", settings.ResponseDeadline, loaded.ResponseDeadline)
	}
}

func TestWorkerRegistrySubmissionWritePaths(t *testing.T) {
	registry, err := NewWorkerRegistry(newTestDB(t))
	if err != nil {
		t.Fatalf("new worker registry: %v", err)
	}

	requestedAt := time.Now().UTC()
	submission := domainrealtime.RegistrationSubmission{
		SubmissionID: "sub-1",
		WorkerID:     "worker-1",
		RequestedAt:  requestedAt,
		ExpiresAt:    requestedAt.Add(5 * time.Minute),
		Status:       domainrealtime.RegistrationStatusPending,
		Capabilities: []domainrealtime.Capability{{
			Contract:     domainrealtime.ContractWorkerRegistry,
			Version:      "1.0.0",
			SubContracts: []domainrealtime.SubContract{domainrealtime.SubContractHeartbeatRequest},
		}},
	}

	if _, err := registry.CreateRegistrationSubmission(context.Background(), submission); err != nil {
		t.Fatalf("create submission: %v", err)
	}

	pending, err := registry.ListPendingRegistrationSubmissions(context.Background(), 10)
	if err != nil {
		t.Fatalf("list pending submissions: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected one pending submission, got %d", len(pending))
	}

	resolvedAt := requestedAt.Add(2 * time.Minute)
	resolved, err := registry.ResolveRegistrationSubmission(context.Background(), "sub-1", domainrealtime.RegistrationStatusAccepted, nil, resolvedAt)
	if err != nil {
		t.Fatalf("resolve submission: %v", err)
	}
	if resolved.Status != domainrealtime.RegistrationStatusAccepted {
		t.Fatalf("expected accepted status, got %s", resolved.Status)
	}
	if resolved.ResolvedAt.Unix() != resolvedAt.Unix() {
		t.Fatalf("expected resolved_at %s, got %s", resolvedAt, resolved.ResolvedAt)
	}

	pending, err = registry.ListPendingRegistrationSubmissions(context.Background(), 10)
	if err != nil {
		t.Fatalf("list pending submissions after resolve: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected no pending submissions after resolve, got %d", len(pending))
	}

	submission2 := submission
	submission2.SubmissionID = "sub-2"
	submission2.RequestedAt = requestedAt.Add(3 * time.Minute)
	submission2.ExpiresAt = submission2.RequestedAt.Add(5 * time.Minute)
	if _, err := registry.CreateRegistrationSubmission(context.Background(), submission2); err != nil {
		t.Fatalf("create submission 2: %v", err)
	}

	revokedAt := submission2.RequestedAt.Add(1 * time.Minute)
	revoked, err := registry.RevokeRegistrationSubmission(context.Background(), "sub-2", "security policy", revokedAt)
	if err != nil {
		t.Fatalf("revoke submission: %v", err)
	}
	if revoked.Status != domainrealtime.RegistrationStatusRevoked {
		t.Fatalf("expected revoked status, got %s", revoked.Status)
	}
	if len(revoked.RejectReasons) != 1 || revoked.RejectReasons[0] != "security policy" {
		t.Fatalf("expected revoke reason to persist, got %#v", revoked.RejectReasons)
	}
}

func TestWorkerRegistryReturnsEpochMismatchOnStaleWrite(t *testing.T) {
	registry, err := NewWorkerRegistry(newTestDB(t))
	if err != nil {
		t.Fatalf("new worker registry: %v", err)
	}

	now := time.Now().UTC()
	if _, err := registry.Register(context.Background(), "worker-1", []taskengine.JobKind{taskengine.JobKindSCMWorkflow}, now, now.Add(30*time.Second)); err != nil {
		t.Fatalf("register worker: %v", err)
	}

	_, err = registry.TouchHeartbeat(context.Background(), "worker-1", 999, now.Add(1*time.Minute), now.Add(2*time.Minute))
	if !errors.Is(err, applicationworker.ErrEpochMismatch) {
		t.Fatalf("expected epoch mismatch error, got %v", err)
	}
}
