package memory

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/topic"
)

type memoryTopicRepository struct {
	mutex  *sync.RWMutex
	topics map[string]domain.Topic
}

func NewMemoryTopicRepository() topic.Repository {
	return &memoryTopicRepository{
		mutex:  new(sync.RWMutex),
		topics: make(map[string]domain.Topic),
	}
}

func (repo *memoryTopicRepository) Update(ctx context.Context, u *url.URL, update topic.UpdateFunc) error {
	tx, err := repo.Get(ctx, u)
	if err != nil {
		return fmt.Errorf("cannot find updating topic: %w", err)
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	result, err := update(tx)
	if err != nil {
		return fmt.Errorf("cannot update topic: %w", err)
	}

	repo.topics[u.String()] = *result

	return nil
}

func (repo *memoryTopicRepository) Create(ctx context.Context, u *url.URL, t domain.Topic) error {
	_, err := repo.Get(ctx, u)
	if err != nil && !errors.Is(err, topic.ErrNotExist) {
		return fmt.Errorf("cannot get topic: %w", err)
	}

	if err == nil {
		return topic.ErrExist
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.topics[u.String()] = t

	return nil
}

func (repo *memoryTopicRepository) Get(ctx context.Context, u *url.URL) (*domain.Topic, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if out, ok := repo.topics[u.String()]; ok {
		return &out, nil
	}

	return nil, topic.ErrNotExist
}

func (repo *memoryTopicRepository) Fetch(_ context.Context) ([]domain.Topic, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	out := make([]domain.Topic, 0)

	for _, t := range repo.topics {
		out = append(out, t)
	}

	return out, nil
}
