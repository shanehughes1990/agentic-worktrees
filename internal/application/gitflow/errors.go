package gitflow

import (
	"fmt"
	"strings"
)

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
		classProvider, ok := current.(interface{ FailureClass() string })
		if ok {
			class := FailureClass(strings.ToLower(strings.TrimSpace(classProvider.FailureClass())))
			if class == FailureClassTransient || class == FailureClassTerminal {
				target.Class = class
				target.Err = current
				return true
			}
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

func IsTransientInfrastructureFailure(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if message == "" {
		return false
	}
	transientIndicators := []string{
		"startup probe failed",
		"signal: killed",
		"context deadline exceeded",
		"timeout",
		"temporarily unavailable",
	}
	for _, indicator := range transientIndicators {
		if strings.Contains(message, indicator) {
			return true
		}
	}
	return false
}
