package usecase_test

import (
	"context"
	"testing"

	"source.toby3d.me/toby3d/hub/internal/domain"
	subscriptionmemoryrepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/memory"
	"source.toby3d.me/toby3d/hub/internal/subscription/usecase"
	topicmemoryrepo "source.toby3d.me/toby3d/hub/internal/topic/repository/memory"
)

func TestSubscriptionUseCase_Subscribe(t *testing.T) {
	t.Parallel()

	subscription := domain.TestSubscription(t, "https://example.com/")
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()

	ucase := usecase.NewSubscriptionUseCase(subscriptions, topics)

	ok, err := ucase.Subscribe(context.Background(), *subscription)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}

	if _, err := subscriptions.Get(context.Background(), subscription.SUID()); err != nil {
		t.Fatal(err)
	}

	t.Run("resubscribe", func(t *testing.T) {
		t.Parallel()

		ok, err := ucase.Subscribe(context.Background(), *subscription)
		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Errorf("want %t, got %t", true, ok)
		}
	})
}

func TestSubscriptionUseCase_Unsubscribe(t *testing.T) {
	t.Parallel()

	subscription := domain.TestSubscription(t, "https://example.com/")
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()

	if err := subscriptions.Create(context.Background(), subscription.SUID(), *subscription); err != nil {
		t.Fatal(err)
	}

	ok, err := usecase.NewSubscriptionUseCase(subscriptions, topics).
		Unsubscribe(context.Background(), *subscription)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}

	if _, err := subscriptions.Get(context.Background(), subscription.SUID()); err == nil {
		t.Error("want error, got nil")
	}
}
