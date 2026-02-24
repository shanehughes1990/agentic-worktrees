package gitflow

import "fmt"

type FailureClass string

const (
	FailureClassTransient FailureClass = "transient"
	FailureClassTerminal  FailureClass = "terminal"
)

type ClassifiedError struct {
	Class FailureClass
	Err   error
}

func (err *ClassifiedError) Error() string {
	if err == nil || err.Err == nil {
		return "classified error"
	}
	return err.Err.Error()
}

func (err *ClassifiedError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func WrapTerminal(err error) error {
	if err == nil {
		return nil
	}
	return &ClassifiedError{Class: FailureClassTerminal, Err: err}
}

func WrapTransient(err error) error {
	if err == nil {
		return nil
	}
	return &ClassifiedError{Class: FailureClassTransient, Err: err}
}

func IsTerminalFailure(err error) bool {
	if err == nil {
		return false
	}
	classifiedErr := &ClassifiedError{}
	if !asClassifiedError(err, classifiedErr) {
		return false
	}
	return classifiedErr.Class == FailureClassTerminal
}

func asClassifiedError(err error, target *ClassifiedError) bool {
	if err == nil || target == nil {
		return false
	}
	current := err
	for current != nil {
		classifiedErr, ok := current.(*ClassifiedError)
		if ok {
			target.Class = classifiedErr.Class
			target.Err = classifiedErr.Err
			return true
		}
		type unwrapper interface{ Unwrap() error }
		wrapped, ok := current.(unwrapper)
		if !ok {
			return false
		}
		current = wrapped.Unwrap()
	}
	return false
}

func EnsureClassified(err error, defaultClass FailureClass) error {
	if err == nil {
		return nil
	}
	classifiedErr := &ClassifiedError{}
	if asClassifiedError(err, classifiedErr) {
		return err
	}
	if defaultClass == FailureClassTerminal {
		return WrapTerminal(fmt.Errorf("%w", err))
	}
	return WrapTransient(fmt.Errorf("%w", err))
}
