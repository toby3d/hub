package domain

import "net/url"

type QueryAdder interface {
	AddQuery(q url.Values)
}
