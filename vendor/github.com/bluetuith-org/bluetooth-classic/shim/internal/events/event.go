//go:build !linux

package events

import (
	"bytes"
	"errors"
	"io"

	"github.com/ugorji/go/codec"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/serde"
)

type RawEvents interface {
	bluetooth.Events | AuthEventData
}

// ServerEvent describes a raw event that was sent from the server.
type ServerEvent struct {
	EventId     bluetooth.EventID     `json:"event_id,omitempty"`
	EventAction bluetooth.EventAction `json:"event_action"`
	Event       codec.Raw             `json:"event"`
}

// Unmarshal unmarshals a 'ServerEvent' to a bluetooth event.
func Unmarshal[T bluetooth.Events](ev ServerEvent) (T, error) {
	var event T

	return event, UnmarshalRawEvent(ev, &event)
}

func UnmarshalRawEvent[T RawEvents](ev ServerEvent, marshalTo *T) error {
	var read int

	scanner := bytes.NewReader(ev.Event)
	for {
		c, err := scanner.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		if c == ':' {
			break
		}

		read++
	}

	if read+1 >= len(ev.Event) {
		return errors.New("adapter update decode error")
	}

	ev.Event = ev.Event[read+1:]

	return serde.UnmarshalJson(ev.Event, marshalTo)
}
