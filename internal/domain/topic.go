package domain

import (
	"net/url"
	"testing"
	"time"

	"source.toby3d.me/toby3d/hub/internal/common"
)

type Topic struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Self        *url.URL
	ContentType string
	Content     []byte
}

func TestTopic(tb testing.TB) *Topic {
	tb.Helper()

	now := time.Now().UTC().Add(-1 * time.Hour)

	return &Topic{
		CreatedAt:   now,
		UpdatedAt:   now,
		Self:        &url.URL{Scheme: "https", Host: "example.com", Path: "/"},
		ContentType: "text/html",
		Content:     []byte("hello, world"),
	}
}

func (t Topic) AddQuery(q url.Values) {
	q.Add(common.HubTopic, t.Self.String())
}

func (t Topic) Equal(target Topic) bool {
	return t.Self.String() == target.Self.String()
}

func (t Topic) String() string {
	return t.Self.String()
}
