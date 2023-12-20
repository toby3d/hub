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
		// TODO(toby3d): search by URL prefix for publish every topic on
		// domain or it's directory.
		Fetch(ctx context.Context) ([]domain.Topic, error)
		Get(ctx context.Context, u *url.URL) (*domain.Topic, error)
		// TODO(toby3d): Delete(ctx context.Context, u *url.URL) (bool, error)
	}
)

var (
	ErrExist    = errors.New("topic already exists")
	ErrNotExist = errors.New("topic does not exist")
)
