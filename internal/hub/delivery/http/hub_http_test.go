package http_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/text/language"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	delivery "source.toby3d.me/toby3d/hub/internal/hub/delivery/http"
	hubucase "source.toby3d.me/toby3d/hub/internal/hub/usecase"
	subscriptionmemoryrepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/memory"
	subscriptionucase "source.toby3d.me/toby3d/hub/internal/subscription/usecase"
	topicmemoryrepo "source.toby3d.me/toby3d/hub/internal/topic/repository/memory"
	topicucase "source.toby3d.me/toby3d/hub/internal/topic/usecase"
)

func TestHandler_ServeHTTP_Subscribe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Query().Get(common.HubChallenge))
	}))
	t.Cleanup(srv.Close)

	in := domain.TestSubscription(t, srv.URL+"/lipsum")
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	hub := hubucase.NewHubUseCase(topics, subscriptions, srv.Client(), &url.URL{
		Scheme: "https",
		Host:   "hub.exmaple.com",
		Path:   "/",
	})

	payload := make(url.Values)
	domain.ModeSubscribe.AddQuery(payload)
	in.AddQuery(payload)

	req := httptest.NewRequest(http.MethodPost, "https://hub.example.com/", strings.NewReader(payload.Encode()))
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationFormCharsetUTF8)

	w := httptest.NewRecorder()
	delivery.NewHandler(delivery.NewHandlerParams{
		Hub:           hub,
		Subscriptions: subscriptionucase.NewSubscriptionUseCase(subscriptions, topics, srv.Client()),
		Topics:        topicucase.NewTopicUseCase(topics, srv.Client()),
		Matcher:       language.NewMatcher([]language.Tag{language.English}),
		Name:          "WebSub",
	}).ServeHTTP(w, req)

	resp := w.Result()

	if expect := http.StatusAccepted; resp.StatusCode != expect {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, expect)
	}
}

func TestHandler_ServeHTTP_Unsubscribe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
		fmt.Fprint(w, r.URL.Query().Get(common.HubChallenge))
	}))
	t.Cleanup(srv.Close)

	in := domain.TestSubscription(t, srv.URL+"/lipsum")
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()
	topics := topicmemoryrepo.NewMemoryTopicRepository()

	if err := subscriptions.Create(context.Background(), in.SUID(), *in); err != nil {
		t.Fatal(err)
	}

	hub := hubucase.NewHubUseCase(topics, subscriptions, srv.Client(), &url.URL{
		Scheme: "https",
		Host:   "hub.exmaple.com",
		Path:   "/",
	})

	payload := make(url.Values)
	domain.ModeUnsubscribe.AddQuery(payload)
	in.AddQuery(payload)

	req := httptest.NewRequest(http.MethodPost, "https://hub.example.com/", strings.NewReader(payload.Encode()))
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationFormCharsetUTF8)

	w := httptest.NewRecorder()
	delivery.NewHandler(delivery.NewHandlerParams{
		Hub:           hub,
		Subscriptions: subscriptionucase.NewSubscriptionUseCase(subscriptions, topics, srv.Client()),
		Topics:        topicucase.NewTopicUseCase(topics, srv.Client()),
		Matcher:       language.NewMatcher([]language.Tag{language.English}),
		Name:          "WebSub",
	}).ServeHTTP(w, req)

	resp := w.Result()

	if expect := http.StatusAccepted; resp.StatusCode != expect {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, expect)
	}
}
