package types

import (
	"fmt"
	"strings"
)

type RunID string
type TaskID string
type JobID string

type Correlation struct {
	RunID  RunID
	TaskID TaskID
	JobID  JobID
}

func (id RunID) String() string {
	return string(id)
}

func (id TaskID) String() string {
	return string(id)
}

func (id JobID) String() string {
	return string(id)
}

func ValidateNonEmpty[T ~string](label string, value T) error {
	if strings.TrimSpace(string(value)) == "" {
		return fmt.Errorf("%s cannot be empty", label)
	}
	return nil
}

func (c Correlation) Validate() error {
	if err := ValidateNonEmpty("run_id", c.RunID); err != nil {
		return err
	}
	if err := ValidateNonEmpty("task_id", c.TaskID); err != nil {
		return err
	}
	if err := ValidateNonEmpty("job_id", c.JobID); err != nil {
		return err
	}
	return nil
}
