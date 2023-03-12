package http_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/language"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	delivery "source.toby3d.me/toby3d/hub/internal/hub/delivery/http"
	ucase "source.toby3d.me/toby3d/hub/internal/hub/usecase"
	"source.toby3d.me/toby3d/hub/internal/subscription"
	subscriptionmemoryrepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/memory"
)

func TestHandler_ServeHTTP_Subscribe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Query().Get(common.HubChallenge))
	}))
	t.Cleanup(srv.Close)

	in := domain.TestSubscription(t, srv.URL+"/lipsum")
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()
	hub := ucase.NewHubUseCase(subscriptions, srv.Client(), &url.URL{Scheme: "https", Host: "hub.exmaple.com"})

	payload := make(url.Values)
	domain.ModeSubscribe.AddQuery(payload)
	in.AddQuery(payload)

	req := httptest.NewRequest(http.MethodPost, "https://hub.example.com/",
		strings.NewReader(payload.Encode()))
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationFormCharsetUTF8)

	w := httptest.NewRecorder()
	delivery.NewHandler(delivery.NewHandlerParams{
		Hub:           hub,
		Subscriptions: subscriptions,
		Matcher:       language.NewMatcher([]language.Tag{language.English}),
		Name:          "hub",
	}).ServeHTTP(w, req)

	resp := w.Result()

	if expect := http.StatusAccepted; resp.StatusCode != expect {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, expect)
	}

	out, err := subscriptions.Get(context.Background(), in.SUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(out, in, cmp.AllowUnexported(domain.Secret{}, domain.Callback{}, domain.Topic{},
		domain.LeaseSeconds{})); diff != "" {
		t.Error(diff)
	}
}

func TestHandler_ServeHTTP_Unsubscribe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Query().Get(common.HubChallenge))
	}))
	t.Cleanup(srv.Close)

	in := domain.TestSubscription(t, srv.URL+"/lipsum")
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()

	if err := subscriptions.Create(context.Background(), in.SUID(), *in); err != nil {
		t.Fatal(err)
	}

	hub := ucase.NewHubUseCase(subscriptions, srv.Client(), &url.URL{Scheme: "https", Host: "hub.exmaple.com"})

	payload := make(url.Values)
	domain.ModeUnsubscribe.AddQuery(payload)
	in.AddQuery(payload)

	req := httptest.NewRequest(http.MethodPost, "https://hub.example.com/",
		strings.NewReader(payload.Encode()))
	req.Header.Set(common.HeaderContentType, common.MIMEApplicationFormCharsetUTF8)

	w := httptest.NewRecorder()
	delivery.NewHandler(delivery.NewHandlerParams{
		Hub:           hub,
		Subscriptions: subscriptions,
		Matcher:       language.NewMatcher([]language.Tag{language.English}),
		Name:          "hub",
	}).ServeHTTP(w, req)

	resp := w.Result()

	if expect := http.StatusAccepted; resp.StatusCode != expect {
		t.Errorf("%s %s = %d, want %d", req.Method, req.RequestURI, resp.StatusCode, expect)
	}

	if _, err := subscriptions.Get(context.Background(), in.SUID()); !errors.Is(err, subscription.ErrNotExist) {
		t.Errorf("want %s, got %s", subscription.ErrNotExist, err)
	}
}
