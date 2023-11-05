package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"

	"source.toby3d.me/toby3d/hub/internal/common"
)

type Challenge struct {
	challenge string
}

func NewChallenge(length uint8) (*Challenge, error) {
	src := make([]byte, length)
	if _, err := rand.Read(src); err != nil {
		return nil, fmt.Errorf("cannot create a new challenge: %w", err)
	}

	return &Challenge{challenge: base64.URLEncoding.EncodeToString(src)}, nil
}

func (c Challenge) AddQuery(q url.Values) {
	q.Add(common.HubChallenge, string(c.challenge))
}

func (c Challenge) Equal(target string) bool {
	return c.challenge == target
}

func (c Challenge) String() string {
	return string(c.challenge)
}
