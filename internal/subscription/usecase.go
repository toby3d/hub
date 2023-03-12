package subscription

import (
	"context"

	"source.toby3d.me/toby3d/hub/internal/domain"
)

type UseCase interface {
	Fetch(ctx context.Context, topic domain.Topic) ([]domain.Subscription, error)
}
