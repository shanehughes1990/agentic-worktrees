package taskengine

import "errors"

var (
	ErrInvalidEnqueueRequest         = errors.New("taskengine: invalid enqueue request")
	ErrInvalidRemoteExecutionRequest = errors.New("taskengine: invalid remote execution request")
)
