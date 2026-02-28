package taskengine

import "time"

func DefaultPolicies() map[JobKind]JobPolicy {
	return map[JobKind]JobPolicy{
		JobKindIngestionAgent: {
			DefaultQueue:          "ingestion",
			RequireIdempotencyKey: true,
			RequireUniqueFor:      true,
			DefaultUniqueFor:      2 * time.Hour,
			DefaultTimeout:        5 * time.Minute,
			DefaultMaxRetry:       2,
		},
		JobKindAgentWorkflow: {
			DefaultQueue:          "agent",
			RequireIdempotencyKey: true,
			RequireUniqueFor:      true,
			DefaultUniqueFor:      2 * time.Hour,
			DefaultTimeout:        10 * time.Minute,
			DefaultMaxRetry:       3,
		},
		JobKindSCMWorkflow: {
			DefaultQueue:          "scm",
			RequireIdempotencyKey: true,
			RequireUniqueFor:      true,
			DefaultUniqueFor:      2 * time.Hour,
			DefaultTimeout:        10 * time.Minute,
			DefaultMaxRetry:       3,
		},
	}
}
