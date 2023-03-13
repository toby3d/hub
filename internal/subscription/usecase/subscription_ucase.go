package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/subscription"
	"source.toby3d.me/toby3d/hub/internal/topic"
)

type subscriptionUseCase struct {
	topics        topic.Repository
	subscriptions subscription.Repository
	client        *http.Client
}

func NewSubscriptionUseCase(subs subscription.Repository, tops topic.Repository, c *http.Client) subscription.UseCase {
	return &subscriptionUseCase{
		subscriptions: subs,
		topics:        tops,
		client:        c,
	}
}

func (ucase *subscriptionUseCase) Subscribe(ctx context.Context, s domain.Subscription) (bool, error) {
	now := time.Now().UTC().Round(time.Second)

	if _, err := ucase.topics.Get(context.Background(), s.Topic); err != nil {
		if !errors.Is(err, topic.ErrNotExist) {
			return false, fmt.Errorf("cannot check subscription topic: %w", err)
		}

		resp, err := ucase.client.Get(s.Topic.String())
		if err != nil {
			return false, fmt.Errorf("cannot fetch a new topic subscription content: %w", err)
		}

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("cannot read a new topic subscription content: %w", err)
		}

		if err = ucase.topics.Create(ctx, s.Topic, domain.Topic{
			CreatedAt:   now,
			UpdatedAt:   now,
			Self:        s.Topic,
			ContentType: resp.Header.Get(common.HeaderContentType),
			Content:     content,
		}); err != nil {
			return false, fmt.Errorf("cannot create topic for subsciption: %w", err)
		}
	}

	if err := ucase.subscriptions.Create(ctx, s.SUID(), domain.Subscription{
		CreatedAt: now,
		UpdatedAt: now,
		SyncedAt:  now,
		ExpiredAt: s.ExpiredAt,
		Callback:  s.Callback,
		Topic:     s.Topic,
		Secret:    s.Secret,
	}); err != nil {
		if !errors.Is(err, subscription.ErrExist) {
			return false, fmt.Errorf("cannot create a new subscription: %w", err)
		}

		if err = ucase.subscriptions.Update(ctx, s.SUID(), func(tx *domain.Subscription) (*domain.Subscription,
			error,
		) {
			tx.UpdatedAt = now
			tx.ExpiredAt = now.Add(time.Duration(s.LeaseSeconds()) * time.Second)
			tx.Secret = s.Secret

			return tx, nil
		}); err != nil {
			return false, fmt.Errorf("cannot resubscribe existing subscription: %w", err)
		}
	}

	return true, nil
}

func (ucase *subscriptionUseCase) Unsubscribe(ctx context.Context, s domain.Subscription) (bool, error) {
	if err := ucase.subscriptions.Delete(ctx, s.SUID()); err != nil {
		return false, fmt.Errorf("cannot unsubscribe: %w", err)
	}

	return true, nil
}
