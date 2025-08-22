package fmsg

import (
	"errors"
	"fmt"
	"strings"
)

// Issue describes an error message that is intended for an end-user to read it.
// It's named "issue" because this word feels like a "non-technical" alternative
// to the word "error" which is 1. overused and 2. typically used mainly in tech
// to describe internal tech issues. Using a different word helps differentiate.
type Issue = string

type withMessage struct {
	underlying error
	internal   string
	external   string
}

// Wrap wraps an error with an internal and an external message. The internal
// message is just the same as any error wrapping library but the external error
// message is intended for display to the end-user. It's recommended to use full
// punctuation and grammar and end the message with a period.
func Wrap(err error, internal, external string) error {
	if err == nil {
		return nil
	}

	return &withMessage{
		err, internal, external,
	}
}

// With implements the Fault Wrapper interface and calls `Wrap` with just the
// internal error message. See `Wrap` for more details.
func With(internal string) func(error) error {
	return func(err error) error {
		return Wrap(err, internal, "")
	}
}

// Withf is a shorthand for With(fmt.Sprintf()). See `With` for more details.
func Withf(internal string, va ...any) func(error) error {
	return func(err error) error {
		return Wrap(err, fmt.Sprintf(internal, va...), "")
	}
}

// WithDesc allows an additional description message to be set. These messages
// are accessible by calling `GetIssue` on an error chain. These descriptions
// are intended to be exposed to end-users as error/diagnostic messages.
func WithDesc(internal, description string) func(error) error {
	return func(err error) error {
		return Wrap(err, internal, description)
	}
}

// Error satisfies the error interface by returning the internal error message.
func (e *withMessage) Error() string { return e.internal }

// Unwrap satisfies the errors unwrap interface.
func (e *withMessage) Unwrap() error { return e.underlying }

// GetIssue returns a space-joined string of all end-user issue messages in the
// error chain. This message can then be displayed/sent to end users.
func GetIssue(err error) Issue {
	return Issue(strings.Join(GetIssues(err), " "))
}

// GetIssues returns all end-user intended messages in the input error chain.
func GetIssues(err error) []Issue {
	p := []Issue{}

	for err != nil {
		if wm, ok := err.(*withMessage); ok {
			if wm.external != "" {
				p = append(p, wm.external)
			}
		}

		err = errors.Unwrap(err)
	}

	return p
}
