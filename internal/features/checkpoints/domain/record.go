package domain

import "time"

type Record struct {
	RunID     string    `json:"run_id"`
	TaskID    string    `json:"task_id"`
	JobID     string    `json:"job_id"`
	State     string    `json:"state"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
