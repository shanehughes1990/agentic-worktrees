package failures

import "errors"

type Class string

const (
	ClassUnknown   Class = "unknown"
	ClassTransient Class = "transient"
	ClassTerminal  Class = "terminal"
)

type ClassifiedError struct {
	class Class
	err   error
}

func (classifiedError ClassifiedError) Error() string {
	if classifiedError.err == nil {
		return string(classifiedError.class)
	}
	return classifiedError.err.Error()
}

func (classifiedError ClassifiedError) Unwrap() error {
	return classifiedError.err
}

func (classifiedError ClassifiedError) Class() Class {
	if classifiedError.class == "" {
		return ClassUnknown
	}
	return classifiedError.class
}

func Wrap(err error, class Class) error {
	if err == nil {
		return nil
	}
	if class == "" {
		class = ClassUnknown
	}
	var existing ClassifiedError
	if errors.As(err, &existing) {
		return err
	}
	return ClassifiedError{class: class, err: err}
}

func WrapTransient(err error) error {
	return Wrap(err, ClassTransient)
}

func WrapTerminal(err error) error {
	return Wrap(err, ClassTerminal)
}

func ClassOf(err error) Class {
	if err == nil {
		return ClassUnknown
	}
	var classifiedError ClassifiedError
	if errors.As(err, &classifiedError) {
		return classifiedError.Class()
	}
	return ClassUnknown
}

func IsClass(err error, class Class) bool {
	return ClassOf(err) == class
}
