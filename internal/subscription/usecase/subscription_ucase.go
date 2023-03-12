package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/subscription"
)

type subscriptionUseCase struct {
	subscriptions subscription.Repository
}

func NewSubscriptionUseCase(subscriptions subscription.Repository) subscription.UseCase {
	return &subscriptionUseCase{
		subscriptions: subscriptions,
	}
}

func (ucase *subscriptionUseCase) Fetch(ctx context.Context, topic domain.Topic) ([]domain.Subscription, error) {
	out, err := ucase.subscriptions.Fetch(ctx, topic)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch subscriptions for topic: %w", err)
	}

	return out, nil
}
