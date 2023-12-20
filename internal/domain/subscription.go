package domain

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"source.toby3d.me/toby3d/hub/internal/common"
)

// Subscription is a unique relation to a topic by a subscriber that indicates
// it should receive updates for that topic.
type Subscription struct {
	// First creation datetime
	CreatedAt time.Time

	// Last updating datetime
	UpdatedAt time.Time

	// Datetime when subscription must be deleted
	ExpiredAt time.Time

	// Datetime synced with topic updating time
	SyncedAt time.Time

	Callback *url.URL
	Topic    *url.URL

	Secret Secret
}

func (s Subscription) AddQuery(q url.Values) {
	s.Secret.AddQuery(q)
	q.Add(common.HubTopic, s.Topic.String())
	q.Add(common.HubCallback, s.Callback.String())
	q.Add(common.HubLeaseSeconds, strconv.FormatFloat(s.LeaseSeconds(), 'g', 0, 64))
}

func (s Subscription) SUID() SUID {
	return SUID{
		topic:    s.Topic,
		callback: s.Callback,
	}
}

func (s Subscription) LeaseSeconds() float64 {
	return s.ExpiredAt.Sub(s.UpdatedAt).Round(time.Second).Seconds()
}

func (s Subscription) Synced(t Topic) bool {
	return s.SyncedAt.Equal(t.UpdatedAt) || s.SyncedAt.After(t.UpdatedAt)
}

func (s Subscription) Expired(ts time.Time) bool {
	return s.ExpiredAt.Before(ts)
}

func TestSubscription(tb testing.TB, callbackUrl string) *Subscription {
	tb.Helper()

	callback, err := url.Parse(callbackUrl)
	if err != nil {
		tb.Fatal(err)
	}

	ts := time.Now().UTC().Round(time.Second)
	secret := TestSecret(tb)

	return &Subscription{
		CreatedAt: ts,
		UpdatedAt: ts,
		ExpiredAt: ts.Add(10 * 24 * time.Hour).Round(time.Second),
		Callback:  callback,
		Topic:     &url.URL{Scheme: "https", Host: "example.com", Path: "/lipsum"},
		Secret:    *secret,
	}
}
