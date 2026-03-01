package taskengine

import "errors"

var (
	ErrInvalidEnqueueRequest                = errors.New("taskengine: invalid enqueue request")
	ErrInvalidRemoteExecutionRequest        = errors.New("taskengine: invalid remote execution request")
	ErrInvalidWorkerCapabilityAdvertisement = errors.New("taskengine: invalid worker capability advertisement")
	ErrInvalidLeaseContract                 = errors.New("taskengine: invalid lease contract")
	ErrInvalidLeaseRenewal                  = errors.New("taskengine: invalid lease renewal")
	ErrInvalidExecutionRecord               = errors.New("taskengine: invalid execution record")
	ErrInvalidDeadLetterRequest             = errors.New("taskengine: invalid dead letter request")
)
