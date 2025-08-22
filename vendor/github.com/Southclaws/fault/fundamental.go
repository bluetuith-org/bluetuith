package fault

import "fmt"

// New creates a new basic fault error.
func New(message string, w ...Wrapper) error {
	f := &fundamental{
		msg:      message,
		location: getLocation(),
	}

	var err error = f
	for _, fn := range w {
		err = fn(err)
	}

	return err
}

// Newf includes formatting specifiers.
func Newf(message string, va ...any) error {
	f := &fundamental{
		msg:      fmt.Sprintf(message, va...),
		location: getLocation(),
	}
	return f
}

type fundamental struct {
	msg      string
	location string
}

func (f *fundamental) Error() string {
	return f.msg
}
