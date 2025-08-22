package ftag

import "errors"

type withKind struct {
	underlying error
	tag        Kind
}

// Kind is a simple string to describe the category of an error.
type Kind string

// Implements all the interfaces for compatibility with the errors ecosystem.

func (e *withKind) Error() string  { return "<ftag>" }
func (e *withKind) Cause() error   { return e.underlying }
func (e *withKind) Unwrap() error  { return e.underlying }
func (e *withKind) String() string { return e.Error() }

// Wrap wraps an error and gives it a distinct tag.
func Wrap(parent error, k Kind) error {
	if parent == nil {
		return nil
	}

	if k == "" {
		return parent
	}

	return &withKind{
		underlying: parent,
		tag:        k,
	}
}

// With implements the Fault Wrapper interface.
func With(k Kind) func(error) error {
	return func(err error) error {
		return Wrap(err, k)
	}
}

// Get extracts the error tag of an error chain. If there's no tag, returns nil.
func Get(err error) Kind {
	if err == nil {
		return None
	}

	for err != nil {
		if f, ok := err.(*withKind); ok {
			return f.tag
		}

		err = errors.Unwrap(err)
	}

	return Internal
}

func GetAll(err error) []Kind {
	if err == nil {
		return nil
	}

	var ks []Kind
	for err != nil {
		if f, ok := err.(*withKind); ok {
			ks = append(ks, f.tag)
		}

		err = errors.Unwrap(err)
	}

	return ks
}

// Common kinds of error:

const (
	None             Kind = ""                  // Empty error.
	Internal         Kind = "INTERNAL"          // Internal errors. This means that some invariants expected by the underlying system have been broken. This error code is reserved for serious errors.
	Cancelled        Kind = "CANCELLED"         // The operation was cancelled, typically by the caller.
	InvalidArgument  Kind = "INVALID_ARGUMENT"  // The client specified an invalid argument.
	NotFound         Kind = "NOT_FOUND"         // Some requested entity was not found.
	AlreadyExists    Kind = "ALREADY_EXISTS"    // The entity that a client attempted to create already exists.
	PermissionDenied Kind = "PERMISSION_DENIED" // The caller does not have permission to execute the specified operation.
	Unauthenticated  Kind = "UNAUTHENTICATED"   // The request does not have valid authentication credentials for the operation.
)
