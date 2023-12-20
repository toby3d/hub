package usecase

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/hub"
	"source.toby3d.me/toby3d/hub/internal/subscription"
	"source.toby3d.me/toby3d/hub/internal/topic"
)

type hubUseCase struct {
	subscriptions subscription.Repository
	topics        topic.Repository
	client        *http.Client
	self          *url.URL
}

const (
	lengthMin = 16
	lengthMax = 32
)

func NewHubUseCase(t topic.Repository, s subscription.Repository, c *http.Client, u *url.URL) hub.UseCase {
	return &hubUseCase{
		client:        c,
		self:          u,
		topics:        t,
		subscriptions: s,
	}
}

func (ucase *hubUseCase) Verify(ctx context.Context, s domain.Subscription, mode domain.Mode) (bool, error) {
	challenge, err := domain.NewChallenge(uint8(lengthMin + rand.Intn(lengthMax-lengthMin)))
	if err != nil {
		return false, fmt.Errorf("cannot generate hub.challenge: %w", err)
	}

	u, _ := url.Parse(s.Callback.String())
	q := u.Query()

	mode.AddQuery(q)
	q.Add(common.HubTopic, s.Topic.String())
	challenge.AddQuery(q)

	if mode == domain.ModeSubscribe {
		q.Add(common.HubLeaseSeconds, strconv.FormatFloat(s.LeaseSeconds(), 'g', 0, 64))
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

	if !challenge.Equal(string(body)) {
		return false, fmt.Errorf("%w: got '%s', want '%s'", hub.ErrChallenge, body, *challenge)
	}

	return true, nil
}

func (ucase *hubUseCase) ListenAndServe(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for ts := range ticker.C {
		ts = ts.Round(time.Second)

		topics, err := ucase.topics.Fetch(ctx)
		if err != nil {
			return fmt.Errorf("cannot fetch topics: %w", err)
		}

		for i := range topics {
			subscriptions, err := ucase.subscriptions.Fetch(ctx, &topics[i])
			if err != nil {
				return fmt.Errorf("cannot fetch subscriptions: %w", err)
			}

			for j := range subscriptions {
				if subscriptions[j].Expired(ts) {
					_, err = ucase.subscriptions.Delete(ctx, subscriptions[j].SUID())
					if err != nil {
						return fmt.Errorf("cannot remove expired subcription: %w", err)
					}

					continue
				}

				if subscriptions[j].Synced(topics[i]) {
					continue
				}

				go ucase.push(ctx, subscriptions[j], topics[i], ts)
			}
		}
	}

	return nil
}

func (ucase *hubUseCase) push(ctx context.Context, s domain.Subscription, t domain.Topic, ts time.Time) (bool, error) {
	req, err := http.NewRequest(http.MethodPost, s.Callback.String(), bytes.NewReader(t.Content))
	if err != nil {
		return false, fmt.Errorf("cannot build request: %w", err)
	}

	req.Header.Set(common.HeaderContentType, t.ContentType)
	req.Header.Set(common.HeaderLink, `<`+ucase.self.String()+`>; rel="hub", <`+s.Topic.String()+`>; rel="self"`)
	setXHubSignatureHeader(req, domain.AlgorithmSHA512, s.Secret, t.Content)

	resp, err := ucase.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("cannot push: %w", err)
	}

	suid := s.SUID()

	// The subscriber's callback URL MAY return an HTTP 410 code to indicate
	// that the subscription has been deleted, and the hub MAY terminate the
	// subscription if it receives that code as a response.
	if resp.StatusCode == http.StatusGone {
		if _, err = ucase.subscriptions.Delete(ctx, suid); err != nil {
			return false, fmt.Errorf("cannot remove deleted subscription: %w", err)
		}

		return true, nil
	}

	// The subscriber's callback URL MUST return an HTTP 2xx response code
	// to indicate a success.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, hub.ErrStatus
	}

	if err = ucase.subscriptions.Update(ctx, suid, func(tx *domain.Subscription) (*domain.Subscription, error) {
		tx.SyncedAt = t.UpdatedAt

		return tx, nil
	}); err != nil {
		return false, fmt.Errorf("cannot sync sybsciption status: %w", err)
	}

	return true, nil
}

func setXHubSignatureHeader(req *http.Request, alg domain.Algorithm, secret domain.Secret, body []byte) {
	if !secret.IsSet() || alg == domain.AlgorithmUnd {
		return
	}

	h := hmac.New(alg.Hash, []byte(secret.String()))
	h.Write(body)

	req.Header.Set(common.HeaderXHubSignature, alg.String()+"="+hex.EncodeToString(h.Sum(nil)))
}
