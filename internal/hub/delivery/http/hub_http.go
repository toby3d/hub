package http

import (
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
	"source.toby3d.me/toby3d/hub/web/template"
)

type (
	Request struct {
		Callback     domain.Callback
		Topic        domain.Topic
		Secret       domain.Secret
		Mode         domain.Mode
		LeaseSeconds domain.LeaseSeconds
	}

	Response struct {
		Mode   domain.Mode
		Topic  domain.Topic
		Reason string
	}

	NewHandlerParams struct {
		Hub           hub.UseCase
		Subscriptions subscription.UseCase
		Matcher       language.Matcher
		Name          string
	}

	Handler struct {
		hub           hub.UseCase
		subscriptions subscription.UseCase
		matcher       language.Matcher
		name          string
	}
)

var DefaultRequestLeaseSeconds = domain.NewLeaseSeconds(uint(time.Duration(time.Hour * 24 * 10).Seconds())) // 10 days

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
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	case http.MethodPost:
		req := NewRequest()
		if err := req.bind(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		s := new(domain.Subscription)
		req.populate(s)

		switch req.Mode {
		case domain.ModeSubscribe:
			h.hub.Subscribe(r.Context(), *s)
		case domain.ModeUnsubscribe:
			h.hub.Unsubscribe(r.Context(), *s)
		case domain.ModePublish:
			go h.hub.Publish(r.Context(), req.Topic)
		}

		w.WriteHeader(http.StatusAccepted)
	case "", http.MethodGet:
		tags, _, _ := language.ParseAcceptLanguage(r.Header.Get(common.HeaderAcceptLanguage))
		tag, _, _ := h.matcher.Match(tags...)
		baseOf := template.NewBaseOf(tag, h.name)

		var page template.Page
		if r.URL.Query().Has(common.HubTopic) {
			topic, err := domain.ParseTopic(r.URL.Query().Get(common.HubTopic))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			subscriptions, err := h.subscriptions.Fetch(r.Context(), *topic)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			page = &template.Topic{
				BaseOf:      baseOf,
				Subscribers: len(subscriptions),
			}
		} else {
			page = &template.Home{BaseOf: baseOf}
		}

		w.Header().Set(common.HeaderContentType, common.MIMETextHTMLCharsetUTF8)
		template.WriteTemplate(w, page)
	}
}

func NewRequest() *Request {
	return &Request{
		Mode:         domain.ModeUnd,
		Callback:     domain.Callback{},
		Secret:       domain.Secret{},
		Topic:        domain.Topic{},
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

	switch r.Mode {
	case domain.ModePublish:
		if !req.PostForm.Has(common.HubURL) {
			return fmt.Errorf("%s parameter for %s %s is required, but not provided", common.HubURL,
				r.Mode, common.HubMode)
		}

		topic, err := domain.ParseTopic(req.PostForm.Get(common.HubURL))
		if err != nil {
			return fmt.Errorf("cannot parse %s: %w", common.HubTopic, err)
		}

		r.Topic = *topic
	case domain.ModeSubscribe, domain.ModeUnsubscribe:
		for _, k := range []string{common.HubTopic, common.HubCallback} {
			if req.PostForm.Has(k) {
				continue
			}

			return fmt.Errorf("%s parameter is required, but not provided", k)
		}

		// NOTE(toby3d): hub.topic
		topic, err := domain.ParseTopic(req.PostForm.Get(common.HubTopic))
		if err != nil {
			return fmt.Errorf("cannot parse %s: %w", common.HubTopic, err)
		}

		r.Topic = *topic

		// NOTE(toby3d): hub.callback
		callback, err := domain.ParseCallback(req.PostForm.Get(common.HubCallback))
		if err != nil {
			return fmt.Errorf("cannot parse %s: %w", common.HubCallback, err)
		}

		r.Callback = *callback

		// NOTE(toby3d): hub.lease_seconds
		if r.Mode != domain.ModeUnsubscribe && req.PostForm.Has(common.HubLeaseSeconds) {
			var ls uint64
			if ls, err = strconv.ParseUint(req.PostForm.Get(common.HubLeaseSeconds), 10, 64); err != nil {
				return fmt.Errorf("cannot parse %s: %w", common.HubLeaseSeconds, err)
			}

			if ls != 0 {
				r.LeaseSeconds = domain.NewLeaseSeconds(uint(ls))
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

func (r Request) populate(s *domain.Subscription) {
	s.Callback = r.Callback
	s.LeaseSeconds = r.LeaseSeconds
	s.Secret = r.Secret
	s.Topic = r.Topic
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
