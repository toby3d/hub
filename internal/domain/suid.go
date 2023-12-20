package domain

import "net/url"

// SUID describes a subscription's unique key is the tuple ([Topic] URL,
// Subscriber [Callback] URL).
type SUID struct {
	topic    *url.URL
	callback *url.URL
}

func NewSSID(topic Topic, callback *url.URL) SUID {
	return SUID{
		topic:    topic.Self,
		callback: callback,
	}
}

func (suid SUID) Topic() *url.URL {
	u, _ := url.Parse(suid.topic.String())

	return u
}

func (suid SUID) Callback() *url.URL {
	u, _ := url.Parse(suid.callback.String())

	return u
}

func (suid SUID) Equal(target SUID) bool {
	return suid.topic == target.topic && suid.callback == target.callback
}

func (suid SUID) GoString() string {
	return "domain.SUID(" + suid.topic.String() + ":" + suid.callback.String() + ")"
}
