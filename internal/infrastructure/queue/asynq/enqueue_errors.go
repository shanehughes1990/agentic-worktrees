package asynq

import (
	"errors"

	"github.com/hibiken/asynq"
)

func isDuplicateEnqueueError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, asynq.ErrDuplicateTask) || errors.Is(err, asynq.ErrTaskIDConflict)
}
