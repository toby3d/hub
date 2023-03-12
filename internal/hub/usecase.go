package hub

import (
	"context"
	"errors"

	"source.toby3d.me/toby3d/hub/internal/domain"
)

type UseCase interface {
	Subscribe(ctx context.Context, subscription domain.Subscription) (bool, error)
	Unsubscribe(ctx context.Context, subscription domain.Subscription) (bool, error)
	Publish(ctx context.Context, t domain.Topic) error
}

var (
	ErrStatus    = errors.New("subscriber replied with a non 2xx status")
	ErrNotFound  = errors.New("subscriber denied verification, responding with a 404 status")
	ErrChallenge = errors.New("the challenge of the hub and the subscriber do not match")
)
