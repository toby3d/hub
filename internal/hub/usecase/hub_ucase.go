package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/hub"
	"source.toby3d.me/toby3d/hub/internal/subscription"
)

type hubUseCase struct {
	subscriptions subscription.Repository
	client        *http.Client
	self          *url.URL
}

const (
	lengthMin = 16
	lengthMax = 32
)

func NewHubUseCase(subscriptions subscription.Repository, client *http.Client, self *url.URL) hub.UseCase {
	return &hubUseCase{
		subscriptions: subscriptions,
		client:        client,
		self:          self,
	}
}

func (ucase *hubUseCase) Verify(ctx context.Context, s domain.Subscription, mode domain.Mode) (bool, error) {
	challenge, err := domain.NewChallenge(uint8(lengthMin + rand.Intn(lengthMax-lengthMin)))
	if err != nil {
		return false, fmt.Errorf("cannot generate hub.challenge: %w", err)
	}

	u := s.Callback.URL()
	q := u.Query()

	for _, w := range []domain.QueryAdder{mode, s.Topic, challenge} {
		w.AddQuery(q)
	}

	if mode == domain.ModeSubscribe {
		s.LeaseSeconds.AddQuery(q)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return false, fmt.Errorf("cannot build verification request: %w", err)
	}

	resp, err := ucase.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("cannot send verification request: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, hub.ErrNotFound
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, hub.ErrStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("cannot verify subscriber response body: %w", err)
	}

	if !challenge.Equal(body) {
		return false, fmt.Errorf("%w: got '%s', want '%s'", hub.ErrChallenge, body, *challenge)
	}

	return true, nil
}

func (ucase *hubUseCase) Subscribe(ctx context.Context, s domain.Subscription) (bool, error) {
	var err error
	if _, err = ucase.Verify(ctx, s, domain.ModeSubscribe); err != nil {
		return false, fmt.Errorf("cannot validate subscription request: %w", err)
	}

	suid := s.SUID()

	if _, err = ucase.subscriptions.Get(ctx, suid); err != nil {
		if !errors.Is(err, subscription.ErrNotExist) {
			return false, fmt.Errorf("cannot check exists subscriptions: %w", err)
		}

		if err = ucase.subscriptions.Create(ctx, suid, s); err != nil {
			return false, fmt.Errorf("cannot create a new subscription: %w", err)
		}

		return true, nil
	}

	if err = ucase.subscriptions.Update(ctx, suid, func(buf *domain.Subscription) (*domain.Subscription, error) {
		buf.LeaseSeconds = s.LeaseSeconds
		buf.Secret = s.Secret

		return buf, nil
	}); err != nil {
		return false, fmt.Errorf("cannot update subscription: %w", err)
	}

	return false, nil
}

func (ucase *hubUseCase) Unsubscribe(ctx context.Context, s domain.Subscription) (bool, error) {
	var err error
	if _, err = ucase.Verify(ctx, s, domain.ModeUnsubscribe); err != nil {
		return false, fmt.Errorf("cannot validate unsubscription request: %w", err)
	}

	if err = ucase.subscriptions.Delete(ctx, s.SUID()); err != nil {
		return false, fmt.Errorf("cannot remove subscription: %w", err)
	}

	return true, nil
}

func (ucase *hubUseCase) Publish(ctx context.Context, t domain.Topic) error {
	resp, err := ucase.client.Get(t.String())
	if err != nil {
		return fmt.Errorf("cannot fetch topic payload for publishing: %w", err)
	}

	push := domain.Push{ContentType: resp.Header.Get(common.HeaderContentType)}

	canonicalTopic, err := domain.ParseTopic(resp.Request.URL.String())
	if err != nil {
		return fmt.Errorf("cannot parse canonical topic URL: %w", err)
	}

	push.Self = *canonicalTopic

	if push.Content, err = io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("cannot read topic body: %w", err)
	}

	subscriptions, err := ucase.subscriptions.Fetch(ctx, t)
	if err != nil {
		return fmt.Errorf("cannot fetch subscriptions for topic: %w", err)
	}

	for i := range subscriptions {
		ucase.Push(ctx, push, subscriptions[i])
	}

	return nil
}

func (ucase *hubUseCase) Push(ctx context.Context, p domain.Push, s domain.Subscription) (bool, error) {
	req, err := http.NewRequest(http.MethodPost, s.Callback.String(), bytes.NewReader(p.Content))
	if err != nil {
		return false, fmt.Errorf("cannot build push request: %w", err)
	}

	req.Header.Set(common.HeaderContentType, p.ContentType)
	req.Header.Set(common.HeaderLink, `<`+ucase.self.String()+`>; rel="hub", <`+p.Self.String()+`>; rel="self"`)
	p.SetXHubSignatureHeader(req, domain.AlgorithmSHA512, s.Secret)

	resp, err := ucase.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("cannot push: %w", err)
	}

	// The subscriber's callback URL MAY return an HTTP 410 code to indicate that the subscription has been
	// deleted, and the hub MAY terminate the subscription if it receives that code as a response.
	if resp.StatusCode == http.StatusGone {
		if err = ucase.subscriptions.Delete(ctx, s.SUID()); err != nil {
			return false, fmt.Errorf("cannot remove deleted subscription: %w", err)
		}

		return true, nil
	}

	// The subscriber's callback URL MUST return an HTTP 2xx response code to indicate a success.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, hub.ErrStatus
	}

	return true, nil
}
