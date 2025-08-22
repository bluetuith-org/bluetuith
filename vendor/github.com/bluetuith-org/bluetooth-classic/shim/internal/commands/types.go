//go:build !linux

package commands

import (
	"strings"
	"time"

	"github.com/ugorji/go/codec"
)

// The default timeout to stop waiting for a command's result (response from a server).
const CommandReplyTimeout = 30 * time.Second

type (
	// ExecuteFunc describes an external function that is used to execute the command.
	ExecuteFunc func(params []string) (chan CommandResponse, error)

	// OptionMap describes a map of options to a command.
	OptionMap = map[Option]string

	// NoResult describes an empty result.
	NoResult = struct{}

	// OperationID describes a uniquely generated ID that is provided by the server
	// for the lifetime of the command invocation.
	OperationID uint32

	// RequestID describes a unique ID that is attached to the request (to track the status of the invoked command)
	// by the client.
	RequestID int64
)

// Command describes an entire command and its options.
// T is the return value type of the command.
// If T is of type NoResult, it means the command only returns errors, and no other values.
type Command[T any] struct {
	cmd    string
	optmap OptionMap
}

// CommandResponse is the raw response or result for an invoked command sent from
// the server.
type CommandResponse struct {
	Status string `json:"status"`

	OperationId OperationID  `json:"operation_id,omitempty"`
	RequestId   RequestID    `json:"request_id,omitempty"`
	Error       CommandError `json:"error"`
	Data        codec.Raw    `json:"data"`
}

// CommandError describes an error that occurred while invoking the command,
// whcih is sent from the server.
type CommandError struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

// Error returns a string representation of the underlying error.
func (c CommandError) Error() string {
	sb := strings.Builder{}

	sb.WriteString(c.Name)
	sb.WriteString(": ")
	if c.Description == "" {
		sb.WriteString("No information is provided for this error")
	} else {
		sb.WriteString(c.Description)
	}
	sb.WriteString(". ")

	count := 0
	length := len(c.Metadata)
	if length == 0 {
		goto Print
	}

	sb.WriteString("(")
	for _, v := range c.Metadata {
		count++
		sb.WriteString(v)

		if count < length {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")

Print:
	return sb.String()
}

// String returns a string representation of a command and its options.
func (c *Command[T]) String() string {
	sb := strings.Builder{}
	sb.Grow(len(c.cmd) + (len(c.optmap) * 2))

	sb.WriteString(c.cmd)
	for param, value := range c.optmap {
		sb.WriteString(" ")
		sb.WriteString(string(param))
		sb.WriteString(" ")
		sb.WriteString(value)
	}

	return sb.String()
}

// Slice returns a slice of each of the space-separated elements in a command-options string.
func (c *Command[T]) Slice() []string {
	return strings.Split(c.String(), " ")
}

// WithOption appends a single option type and value to the command's option map.
func (c *Command[T]) WithOption(opt Option, value string) *Command[T] {
	if c.optmap == nil {
		c.optmap = make(OptionMap)
	}

	c.optmap[opt] = value

	return c
}

// WithOptions provides a function to append multiple option-value types to the command's option map.
func (c *Command[T]) WithOptions(fn func(OptionMap)) *Command[T] {
	if c.optmap == nil {
		c.optmap = make(OptionMap)
	}

	fn(c.optmap)

	return c
}
