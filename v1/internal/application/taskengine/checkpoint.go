package taskengine

import "strings"

// Checkpoint identifies a completed step bound to an idempotency token so
// retries can safely decide whether execution should resume or re-run.
type Checkpoint struct {
	Step  string `json:"step"`
	Token string `json:"token"`
}

func CheckpointMatches(checkpoint *Checkpoint, step string, token string) bool {
	if checkpoint == nil {
		return false
	}
	return strings.TrimSpace(checkpoint.Step) == strings.TrimSpace(step) &&
		strings.TrimSpace(checkpoint.Token) == strings.TrimSpace(token)
}
