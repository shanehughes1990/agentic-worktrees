package resolvers

import (
	"agentic-orchestrator/internal/interface/graphql/models"
)

func allSupportedSCMOperations() []models.SCMOperation {
	return []models.SCMOperation{
		models.SCMOperationSourceState,
		models.SCMOperationEnsureWorktree,
		models.SCMOperationSyncWorktree,
		models.SCMOperationCleanupWorktree,
		models.SCMOperationEnsureBranch,
		models.SCMOperationSyncBranch,
		models.SCMOperationUpsertPullRequest,
		models.SCMOperationGetPullRequest,
		models.SCMOperationSubmitReview,
		models.SCMOperationCheckMergeReadiness,
		models.SCMOperationMergePullRequest,
	}
}

