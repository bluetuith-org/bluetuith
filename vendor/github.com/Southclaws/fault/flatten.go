package fault

import (
	"errors"
)

// Chain represents an unwound error chain. Each step is a useful error. Errors
// without messages (such as those that only contain other errors) are omitted.
type Chain []Step

// Step represents a location where an error was wrapped by Fault. The location
// is present if the error being wrapped contained stack information and the
// message is present if the underlying error provided a message. Note that not
// all errors provide errors or locations. If both are missing, it's omitted.
type Step struct {
	Location string
	Message  string
}

// Flatten attempts to derive more useful structured information from an error
// chain. If the input is a fault error, the output will contain an easy to use
// error chain list with location information and individual error messages.
func Flatten(err error) Chain {
	if err == nil {
		return nil
	}

	// first, flatten the call tree into an array so it's easier to work with.
	flat := []error{}
	for err != nil {
		flat = append(flat, err)
		err = errors.Unwrap(err)
	}

	lastLocation := ""

	var f Chain
	for i := 0; i < len(flat); i++ {
		err := flat[i]

		var next error
		if i+1 < len(flat) {
			next = flat[i+1]
		}

		switch unwrapped := err.(type) {
		// NOTE: Because fault containers do not have messages, they only
		// exist to contain other errors that actually contain information,
		// store the container's recorded location for usage with the next item.
		case *container:
			if _, ok := next.(*container); ok && unwrapped.location != "" {
				// Having 2 containers back to back can happen if we're using .Wrap without using any wrappers. In that
				// case, we add a Step to avoid losing the location whe the wrapping occurred
				f = append([]Step{{
					Location: unwrapped.location,
					Message:  "",
				}}, f...)
			}
			lastLocation = unwrapped.location

		case *fundamental:
			f = append([]Step{{
				Location: unwrapped.location,
				Message:  err.Error(),
			}}, f...)

			lastLocation = ""

		default:
			message := err.Error()

			// de-duplicate identical error messages
			if next != nil {
				if message == next.Error() {
					continue
				}
			}

			f = append([]Step{{
				Location: lastLocation,
				Message:  message,
			}}, f...)
		}
	}

	return f
}
