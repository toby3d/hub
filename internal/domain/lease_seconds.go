package domain

import (
	"net/url"
	"strconv"

	"source.toby3d.me/toby3d/hub/internal/common"
)

// LeaseSeconds describes a number of seconds for which the subscriber would
// like to have the subscription active, given as a positive decimal integer.
// Hubs MAY choose to respect this value or not, depending on their own
// policies, and MAY set a default value if the subscriber omits the parameter.
// This parameter MAY be present for unsubscription requests and MUST be ignored
// by the hub in that case.
type LeaseSeconds struct {
	leaseSeconds uint
}

func NewLeaseSeconds(raw uint) LeaseSeconds {
	return LeaseSeconds{leaseSeconds: raw}
}

func (ls LeaseSeconds) AddQuery(q url.Values) {
	if ls.leaseSeconds == 0 {
		return
	}

	q.Add(common.HubLeaseSeconds, strconv.FormatUint(uint64(ls.leaseSeconds), 10))
}

func (ls LeaseSeconds) IsZero() bool {
	return ls.leaseSeconds == 0
}
