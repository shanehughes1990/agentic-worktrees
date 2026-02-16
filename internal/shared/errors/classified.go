package errors

import (
	"errors"
	"fmt"
)

type Class string

const (
	ClassUnknown   Class = "unknown"
	ClassTransient Class = "transient"
	ClassTerminal  Class = "terminal"
)

type ClassifiedError struct {
	Class Class
	Op    string
	Err   error
}

func (e ClassifiedError) Error() string {
	if e.Op == "" {
		return fmt.Sprintf("%s: %v", e.Class, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Class, e.Op, e.Err)
}

func (e ClassifiedError) Unwrap() error {
	return e.Err
}

func Wrap(class Class, op string, err error) error {
	if err == nil {
		return nil
	}
	if class == "" {
		class = ClassUnknown
	}
	return ClassifiedError{Class: class, Op: op, Err: err}
}

func Transient(op string, err error) error {
	return Wrap(ClassTransient, op, err)
}

func Terminal(op string, err error) error {
	return Wrap(ClassTerminal, op, err)
}

func ClassOf(err error) Class {
	if err == nil {
		return ClassUnknown
	}
	var classified ClassifiedError
	if errors.As(err, &classified) {
		if classified.Class == "" {
			return ClassUnknown
		}
		return classified.Class
	}
	return ClassUnknown
}

func IsClass(err error, class Class) bool {
	return ClassOf(err) == class
}
