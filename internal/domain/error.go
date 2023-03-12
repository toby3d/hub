package domain

import (
	"fmt"
	"net/url"

	"golang.org/x/xerrors"
)

type Error struct {
	frame  xerrors.Frame
	topic  *url.URL
	reason string
}

func NewError(reason string, topic ...*url.URL) error {
	err := &Error{
		reason: reason,
		frame:  xerrors.Caller(1),
	}

	if len(topic) > 0 {
		err.topic = topic[0]
	}

	return err
}

// Error returns a string representation of the error, satisfying the error
// interface.
func (e Error) Error() string {
	return fmt.Sprint(e)
}

// Format prints the stack as error detail.
func (e Error) Format(state fmt.State, r rune) {
	xerrors.FormatError(e, state, r)
}

// FormatError prints the receiver's error, if any.
func (e Error) FormatError(printer xerrors.Printer) error {
	printer.Print(e.reason)

	if e.topic != nil {
		printer.Printf(" (%s)", e.topic)
	}

	if !printer.Detail() {
		return nil
	}

	e.frame.Format(printer)

	return nil
}
