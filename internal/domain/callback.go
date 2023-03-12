package domain

import (
	"fmt"
	"net/url"

	"source.toby3d.me/toby3d/hub/internal/common"
)

// Callback describes the URL at which a subscriber wishes to receive content
// distribution requests.
type Callback struct {
	callback *url.URL
}

func ParseCallback(str string) (*Callback, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, fmt.Errorf("cannot parse string as callback URL: %w", err)
	}

	return &Callback{callback: u}, nil
}

func (c Callback) AddQuery(q url.Values) {
	q.Add(common.HubCallback, c.callback.String())
}

func (c Callback) URL() *url.URL {
	u, _ := url.Parse(c.callback.String())

	return u
}

func (c Callback) String() string {
	return c.callback.String()
}
