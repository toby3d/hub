package subscription

import (
	"context"

	"source.toby3d.me/toby3d/hub/internal/domain"
)

type UseCase interface {
	Subscribe(ctx context.Context, s domain.Subscription) (bool, error)
	Unsubscribe(ctx context.Context, s domain.Subscription) (bool, error)
}
