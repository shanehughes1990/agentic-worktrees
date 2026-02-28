package taskengine

import "errors"

var (
	ErrInvalidEnqueueRequest                = errors.New("taskengine: invalid enqueue request")
	ErrInvalidWorkerCapabilityAdvertisement = errors.New("taskengine: invalid worker capability advertisement")
)
