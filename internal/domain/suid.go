package domain

// SUID describes a subscription's unique key is the tuple ([Topic] URL,
// Subscriber [Callback] URL).
type SUID struct {
	suid [2]string
}

func NewSSID(topic Topic, callback Callback) SUID {
	return SUID{
		suid: [2]string{topic.topic.String(), callback.callback.String()},
	}
}

func (suid SUID) Equal(target SUID) bool {
	for i := range suid.suid {
		if suid.suid[i] == target.suid[i] {
			continue
		}

		return false
	}

	return true
}

func (suid SUID) GoString() string {
	return "domain.SUID(" + suid.suid[0] + ":" + suid.suid[1] + ")"
}
