package domain

import "net/url"

// SUID describes a subscription's unique key is the tuple ([Topic] URL,
// Subscriber [Callback] URL).
type SUID struct {
	topic    string
	callback string
}

func NewSSID(topic Topic, callback *url.URL) SUID {
	return SUID{
		topic:    topic.Self.String(),
		callback: callback.String(),
	}
}

func (suid SUID) Equal(target SUID) bool {
	return suid.topic == target.topic && suid.callback == target.callback
}

func (suid SUID) GoString() string {
	return "domain.SUID(" + suid.topic + ":" + suid.callback + ")"
}
