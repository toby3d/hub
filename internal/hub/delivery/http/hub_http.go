package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/text/language"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/hub"
	"source.toby3d.me/toby3d/hub/internal/subscription"
	"source.toby3d.me/toby3d/hub/internal/topic"
	"source.toby3d.me/toby3d/hub/web/template"
)

type (
	Request struct {
		Callback     *url.URL
		Topic        *url.URL
		Secret       domain.Secret
		Mode         domain.Mode
		LeaseSeconds float64
	}

	Response struct {
		Mode   domain.Mode
		Reason string
		Topic  domain.Topic
	}

	NewHandlerParams struct {
		Hub           hub.UseCase
		Subscriptions subscription.UseCase
		Topics        topic.UseCase
		Matcher       language.Matcher
		Name          string
	}

	Handler struct {
		hub           hub.UseCase
		subscriptions subscription.UseCase
		topics        topic.UseCase
		matcher       language.Matcher
		name          string
	}
)

var DefaultRequestLeaseSeconds = time.Duration(10 * 24 * time.Hour).Seconds() // 10 days

var (
	ErrHubMode = errors.New(common.HubMode + " MUST be " + domain.ModeSubscribe.String() + " or " +
		domain.ModeUnsubscribe.String())
	ErrHubSecret = errors.New(common.HubSecret + " SHOULD be specified when the request was made over HTTPS")
)

func NewHandler(params NewHandlerParams) *Handler {
	return &Handler{
		hub:           params.Hub,
		matcher:       params.Matcher,
		name:          params.Name,
		subscriptions: params.Subscriptions,
		topics:        params.Topics,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC().Round(time.Second)

	switch r.Method {
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	case http.MethodPost:
		req := NewRequest()

		var err error
		if err = req.bind(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		// TODO(toby3d): send denied ping to callback if it's not accepted by hub

		s := new(domain.Subscription)
		req.populate(s, now)

		switch req.Mode {
		case domain.ModeSubscribe, domain.ModeUnsubscribe:
			if _, err = h.hub.Verify(r.Context(), *s, req.Mode); err != nil {
				r.Clone(context.WithValue(r.Context(), "error", err))

				w.WriteHeader(http.StatusAccepted)

				return
			}

			switch req.Mode {
			case domain.ModeSubscribe:
				_, err = h.subscriptions.Subscribe(r.Context(), *s)
			case domain.ModeUnsubscribe:
				_, err = h.subscriptions.Unsubscribe(r.Context(), *s)
			}
		case domain.ModePublish:
			_, err = h.topics.Publish(r.Context(), req.Topic)
		}

		if err != nil {
			r.Clone(context.WithValue(r.Context(), "error", err))
		}

		w.WriteHeader(http.StatusAccepted)
	case "", http.MethodGet:
		tags, _, _ := language.ParseAcceptLanguage(r.Header.Get(common.HeaderAcceptLanguage))
		tag, _, _ := h.matcher.Match(tags...)

		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		template.WriteTemplate(w, &template.Home{BaseOf: template.NewBaseOf(tag, h.name)})
	}
}

func NewRequest() *Request {
	return &Request{
		Mode:         domain.ModeUnd,
		Callback:     nil,
		Secret:       domain.Secret{},
		Topic:        nil,
		LeaseSeconds: DefaultRequestLeaseSeconds,
	}
}

func (r *Request) bind(req *http.Request) error {
	var err error
	if err = req.ParseForm(); err != nil {
		return fmt.Errorf("cannot parse request form: %w", err)
	}

	if !req.PostForm.Has(common.HubMode) {
		return fmt.Errorf("%s parameter is required, but not provided", common.HubMode)
	}

	// NOTE(toby3d): hub.mode
	if r.Mode, err = domain.ParseMode(req.PostForm.Get(common.HubMode)); err != nil {
		return fmt.Errorf("cannot parse %s: %w", common.HubMode, err)
	}

	// NOTE(toby3d): hub.topic
	if !req.PostForm.Has(common.HubTopic) {
		return fmt.Errorf("%s parameter is required, but not provided", common.HubTopic)
	}

	if r.Topic, err = url.Parse(req.PostForm.Get(common.HubTopic)); err != nil {
		return fmt.Errorf("cannot parse %s: %w", common.HubTopic, err)
	}

	switch r.Mode {
	case domain.ModePublish:
	case domain.ModeSubscribe, domain.ModeUnsubscribe:
		// NOTE(toby3d): hub.callback
		if !req.PostForm.Has(common.HubCallback) {
			return fmt.Errorf("%s parameter is required, but not provided", common.HubCallback)
		}

		if r.Callback, err = url.Parse(req.PostForm.Get(common.HubCallback)); err != nil {
			return fmt.Errorf("cannot parse %s: %w", common.HubCallback, err)
		}

		// NOTE(toby3d): hub.lease_seconds
		if r.Mode != domain.ModeUnsubscribe && req.PostForm.Has(common.HubLeaseSeconds) {
			r.LeaseSeconds, err = strconv.ParseFloat(req.PostForm.Get(common.HubLeaseSeconds), 64)
			if err != nil {
				return fmt.Errorf("cannot parse %s: %w", common.HubLeaseSeconds, err)
			}
		}

		// NOTE(toby3d): hub.secret
		if !req.PostForm.Has(common.HubSecret) {
			if req.TLS != nil {
				return ErrHubSecret
			}

			return nil
		}

		secret, err := domain.ParseSecret(req.PostForm.Get(common.HubSecret))
		if err != nil {
			return fmt.Errorf("cannot parse %s: %w", common.HubSecret, err)
		}

		r.Secret = *secret
	}

	return nil
}

func (r Request) populate(s *domain.Subscription, ts time.Time) {
	s.CreatedAt = ts
	s.UpdatedAt = ts
	s.ExpiredAt = ts.Add(time.Duration(r.LeaseSeconds) * time.Second).Round(time.Second)
	s.Callback = r.Callback
	s.Topic = r.Topic
	s.Secret = r.Secret
}

func NewResponse(t domain.Topic, err error) *Response {
	return &Response{
		Mode:   domain.ModeDenied,
		Topic:  t,
		Reason: err.Error(),
	}
}

func (r *Response) populate(q url.Values) {
	r.Mode.AddQuery(q)
	r.Topic.AddQuery(q)
	q.Add(common.HubReason, r.Reason)
}
