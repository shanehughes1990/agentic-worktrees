package taskengine

import "testing"

func TestDefaultPoliciesIncludesSCMWorkflowPolicy(t *testing.T) {
	policies := DefaultPolicies()
	policy, ok := policies[JobKindSCMWorkflow]
	if !ok {
		t.Fatalf("expected SCM workflow policy")
	}
	if policy.DefaultQueue != "scm" {
		t.Fatalf("expected scm queue, got %q", policy.DefaultQueue)
	}
	if !policy.RequireIdempotencyKey {
		t.Fatalf("expected SCM policy to require idempotency key")
	}
	if !policy.RequireUniqueFor {
		t.Fatalf("expected SCM policy to require unique_for")
	}
	if policy.DefaultMaxRetry != 3 {
		t.Fatalf("expected SCM max retry 3, got %d", policy.DefaultMaxRetry)
	}
}
