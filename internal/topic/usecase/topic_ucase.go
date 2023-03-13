package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/topic"
)

type topicUseCase struct {
	client *http.Client
	topics topic.Repository
}

func NewTopicUseCase(topics topic.Repository, client *http.Client) topic.UseCase {
	return &topicUseCase{
		topics: topics,
		client: client,
	}
}

func (ucase *topicUseCase) Publish(ctx context.Context, u *url.URL) (bool, error) {
	now := time.Now().UTC().Round(time.Second)

	resp, err := ucase.client.Get(u.String())
	if err != nil {
		return false, fmt.Errorf("cannot fetch publishing url: %w", err)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("cannot read topic response body: %w", err)
	}

	if err := ucase.topics.Update(ctx, u, func(tx *domain.Topic) (*domain.Topic, error) {
		tx.Self = resp.Request.URL
		tx.UpdatedAt = now
		tx.Content = content
		tx.ContentType = resp.Header.Get(common.HeaderContentType)

		return tx, nil
	}); err != nil {
		if !errors.Is(err, topic.ErrNotExist) {
			return false, fmt.Errorf("cannot publish exists topic: %w", err)
		}

		if err = ucase.topics.Create(ctx, resp.Request.URL, domain.Topic{
			CreatedAt:   now,
			UpdatedAt:   now,
			Self:        resp.Request.URL,
			ContentType: resp.Header.Get(common.HeaderContentType),
			Content:     content,
		}); err != nil {
			return false, fmt.Errorf("cannot publish a new topic: %w", err)
		}
	}

	return true, nil
}
