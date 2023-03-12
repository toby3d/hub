package domain

import (
	"errors"
	"fmt"
	"net/url"

	"source.toby3d.me/toby3d/hub/internal/common"
)

type Mode struct {
	mode string
}

var (
	ModeUnd         Mode = Mode{mode: ""}            // "und"
	ModeDenied      Mode = Mode{mode: "denied"}      // "denied"
	ModePublish     Mode = Mode{mode: "publish"}     // "publish"
	ModeSubscribe   Mode = Mode{mode: "subscribe"}   // "subscribe"
	ModeUnsubscribe Mode = Mode{mode: "unsubscribe"} // "unsubscribe"
)

var ErrModeSyntax = errors.New("bad mode syntax")

var stringsModes = map[string]Mode{
	ModeDenied.mode:      ModeDenied,
	ModePublish.mode:     ModePublish,
	ModeSubscribe.mode:   ModeSubscribe,
	ModeUnsubscribe.mode: ModeUnsubscribe,
}

func ParseMode(mode string) (Mode, error) {
	if mode, ok := stringsModes[mode]; ok {
		return mode, nil
	}

	return ModeUnd, fmt.Errorf("%w: %s", ErrModeSyntax, mode)
}

func (m *Mode) UnmarshalForm(src []byte) error {
	var err error
	if *m, err = ParseMode(string(src)); err != nil {
		return fmt.Errorf("Mode: %w", err)
	}

	return nil
}

func (m Mode) AddQuery(q url.Values) {
	q.Add(common.HubMode, m.mode)
}

func (m Mode) String() string {
	if m.mode != "" {
		return m.mode
	}

	return common.Und
}

func (m Mode) GoString() string {
	return "domain.Mode(" + m.String() + ")"
}
