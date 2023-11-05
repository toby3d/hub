package topic

import (
	"context"
	"errors"
	"net/url"

	"source.toby3d.me/toby3d/hub/internal/domain"
)

type (
	UpdateFunc func(t *domain.Topic) (*domain.Topic, error)

	Repository interface {
		Create(ctx context.Context, u *url.URL, topic domain.Topic) error
		Update(ctx context.Context, u *url.URL, update UpdateFunc) error
		Fetch(ctx context.Context) ([]domain.Topic, error)
		Get(ctx context.Context, u *url.URL) (*domain.Topic, error)
	}
)

var (
	ErrExist    = errors.New("topic already exists")
	ErrNotExist = errors.New("topic does not exist")
)
