package domain

import (
	"fmt"
	"net/url"

	"source.toby3d.me/toby3d/hub/internal/common"
)

// Topic is a HTTP [RFC7230] (or HTTPS [RFC2818]) resource URL. The unit to
// which one can subscribe to changes.
//
// [RFC7230]: https://tools.ietf.org/html/rfc7230
// [RFC2818]: https://tools.ietf.org/html/rfc2818
type Topic struct {
	topic *url.URL
}

func ParseTopic(str string) (*Topic, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, fmt.Errorf("cannot parse string as topic URL: %w", err)
	}

	return &Topic{topic: u}, nil
}

func (t Topic) AddQuery(q url.Values) {
	q.Add(common.HubTopic, t.topic.String())
}

func (t Topic) Equal(target Topic) bool {
	return t.topic.String() == target.topic.String()
}

func (t Topic) String() string {
	return t.topic.String()
}
