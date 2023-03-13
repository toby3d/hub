package topic

import (
	"context"
	"net/url"
)

type UseCase interface {
	Publish(ctx context.Context, u *url.URL) (bool, error)
}
