package domain

import (
	"math/rand"
	"net/url"
	"testing"
)

// Subscription is a unique relation to a topic by a subscriber that indicates
// it should receive updates for that topic.
type Subscription struct {
	Topic        Topic
	Callback     Callback
	Secret       Secret
	LeaseSeconds LeaseSeconds
}

func (s Subscription) SUID() SUID {
	return NewSSID(s.Topic, s.Callback)
}

func (s Subscription) AddQuery(q url.Values) {
	for _, w := range []QueryAdder{s.Callback, s.Topic, s.LeaseSeconds, s.Secret} {
		w.AddQuery(q)
	}
}

func TestSubscription(tb testing.TB, callbackUrl string) *Subscription {
	tb.Helper()

	callback, err := ParseCallback(callbackUrl)
	if err != nil {
		tb.Fatal(err)
	}

	return &Subscription{
		Topic: Topic{topic: &url.URL{
			Scheme: "https",
			Host:   "example.com",
			Path:   "lipsum",
		}},
		Callback:     *callback,
		Secret:       *TestSecret(tb),
		LeaseSeconds: NewLeaseSeconds(uint(rand.Intn(60))),
	}
}
