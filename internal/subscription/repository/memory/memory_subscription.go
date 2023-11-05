package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/subscription"
)

type memorySubscriptionRepository struct {
	mutex         *sync.RWMutex
	subscriptions map[domain.SUID]domain.Subscription
}

func NewMemorySubscriptionRepository() subscription.Repository {
	return &memorySubscriptionRepository{
		mutex:         new(sync.RWMutex),
		subscriptions: make(map[domain.SUID]domain.Subscription),
	}
}

func (repo *memorySubscriptionRepository) Create(ctx context.Context, suid domain.SUID, s domain.Subscription) error {
	if _, err := repo.Get(ctx, suid); err != nil {
		if !errors.Is(err, subscription.ErrNotExist) {
			return fmt.Errorf("cannot create subscription: %w", err)
		}
	} else {
		return fmt.Errorf("cannot create subscription: %w", subscription.ErrExist)
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.subscriptions[suid] = s

	return nil
}

func (repo *memorySubscriptionRepository) Delete(ctx context.Context, suid domain.SUID) error {
	if _, err := repo.Get(ctx, suid); err != nil {
		if !errors.Is(err, subscription.ErrNotExist) {
			return fmt.Errorf("cannot delete subscription: %w", err)
		}

		return nil
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.subscriptions, suid)

	return nil
}

func (repo *memorySubscriptionRepository) Get(_ context.Context, suid domain.SUID) (*domain.Subscription, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if out, ok := repo.subscriptions[suid]; ok {
		return &out, nil
	}

	return nil, subscription.ErrNotExist
}

func (repo *memorySubscriptionRepository) Fetch(ctx context.Context, t *domain.Topic) ([]domain.Subscription, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	out := make([]domain.Subscription, 0)

	for _, s := range repo.subscriptions {
		if t != nil && t.Self.String() != s.Topic.String() {
			continue
		}

		out = append(out, s)
	}

	return out, nil
}

// Update implements subscription.Repository
func (repo *memorySubscriptionRepository) Update(ctx context.Context, suid domain.SUID, update subscription.UpdateFunc) error {
	in, err := repo.Get(ctx, suid)
	if err != nil {
		return fmt.Errorf("cannot update subscription: %w", err)
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	out, err := update(in)
	if err != nil {
		return fmt.Errorf("cannot update subscription: %w", err)
	}

	repo.subscriptions[suid] = *out

	return nil
}
