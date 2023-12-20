package subscription

import (
	"context"
	"errors"

	"source.toby3d.me/toby3d/hub/internal/domain"
)

type (
	UpdateFunc func(subscription *domain.Subscription) (*domain.Subscription, error)

	Repository interface {
		Create(ctx context.Context, suid domain.SUID, subscription domain.Subscription) error
		Get(ctx context.Context, suid domain.SUID) (*domain.Subscription, error)
		Fetch(ctx context.Context, topic *domain.Topic) ([]domain.Subscription, error)
		Update(ctx context.Context, suid domain.SUID, update UpdateFunc) error
		Delete(ctx context.Context, suid domain.SUID) (bool, error)
	}
)

var (
	ErrNotExist = errors.New("subscription does not exist")
	ErrExist    = errors.New("subscription already exists")
)
