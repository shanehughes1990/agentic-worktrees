package taskengine

import "strings"

// Checkpoint identifies a completed step bound to an idempotency token so
// retries can safely decide whether execution should resume or re-run.
type Checkpoint struct {
	Step  string `json:"step"`
	Token string `json:"token"`
}

// RetryCheckpointContract defines the checkpoint/resume payload shape used when
// retries cross worker boundaries. Legacy flattened fields remain supported.
type RetryCheckpointContract struct {
	ResumeCheckpoint      *Checkpoint `json:"resume_checkpoint,omitempty"`
	CompletedCheckpoint   *Checkpoint `json:"completed_checkpoint,omitempty"`
	ResumeCheckpointStep  string      `json:"resume_checkpoint_step,omitempty"`
	ResumeCheckpointToken string      `json:"resume_checkpoint_token,omitempty"`
}

func (contract RetryCheckpointContract) Checkpoint() *Checkpoint {
	if checkpoint := normalizedCheckpoint(contract.ResumeCheckpoint); checkpoint != nil {
		return checkpoint
	}
	if checkpoint := normalizedCheckpoint(contract.CompletedCheckpoint); checkpoint != nil {
		return checkpoint
	}
	step := strings.TrimSpace(contract.ResumeCheckpointStep)
	token := strings.TrimSpace(contract.ResumeCheckpointToken)
	if step == "" || token == "" {
		return nil
	}
	return &Checkpoint{Step: step, Token: token}
}

func (contract RetryCheckpointContract) Matches(step string, token string) bool {
	return CheckpointMatches(contract.Checkpoint(), step, token)
}

func CheckpointMatches(checkpoint *Checkpoint, step string, token string) bool {
	expectedStep := strings.TrimSpace(step)
	expectedToken := strings.TrimSpace(token)
	if expectedStep == "" || expectedToken == "" {
		return false
	}
	checkpoint = normalizedCheckpoint(checkpoint)
	if checkpoint == nil {
		return false
	}
	return checkpoint.Step == expectedStep && checkpoint.Token == expectedToken
}

func normalizedCheckpoint(checkpoint *Checkpoint) *Checkpoint {
	if checkpoint == nil {
		return nil
	}
	step := strings.TrimSpace(checkpoint.Step)
	token := strings.TrimSpace(checkpoint.Token)
	if step == "" || token == "" {
		return nil
	}
	return &Checkpoint{
		Step:  step,
		Token: token,
	}
}
