package taskengine

import "errors"

var (
	ErrInvalidEnqueueRequest = errors.New("taskengine: invalid enqueue request")
)
