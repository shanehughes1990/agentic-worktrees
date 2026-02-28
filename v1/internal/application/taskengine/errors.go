package taskengine

import "errors"

var (
	ErrInvalidEnqueueRequest                = errors.New("taskengine: invalid enqueue request")
	ErrInvalidWorkerCapabilityAdvertisement = errors.New("taskengine: invalid worker capability advertisement")
	ErrInvalidLeaseContract                 = errors.New("taskengine: invalid lease contract")
	ErrInvalidLeaseRenewal                  = errors.New("taskengine: invalid lease renewal")
)
