package usecase_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"source.toby3d.me/toby3d/hub/internal/common"
	"source.toby3d.me/toby3d/hub/internal/domain"
	hubucase "source.toby3d.me/toby3d/hub/internal/hub/usecase"
	subscriptionmemoryrepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/memory"
	topicmemoryrepo "source.toby3d.me/toby3d/hub/internal/topic/repository/memory"
)

func TestHubUseCase_Verify(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
		fmt.Fprint(w, r.FormValue(common.HubChallenge))
	}))
	t.Cleanup(srv.Close)

	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	subscription := domain.TestSubscription(t, srv.URL)

	ok, err := hubucase.NewHubUseCase(topics, subscriptions, srv.Client(), &url.URL{
		Scheme: "https",
		Host:   "hub.example.com",
		Path:   "/",
	}).Verify(context.Background(), *subscription, domain.ModeSubscribe)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}
}
