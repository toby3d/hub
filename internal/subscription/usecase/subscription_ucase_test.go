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
	subscriptionmemoryrepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/memory"
	"source.toby3d.me/toby3d/hub/internal/subscription/usecase"
	topicmemoryrepo "source.toby3d.me/toby3d/hub/internal/topic/repository/memory"
)

func TestSubscriptionUseCase_Subscribe(t *testing.T) {
	t.Parallel()

	topic := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
		fmt.Fprint(w, "hello, world")
	}))
	t.Cleanup(topic.Close)

	callback := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
		fmt.Fprint(w, "hello, world")
	}))
	t.Cleanup(callback.Close)

	subscription := domain.TestSubscription(t, callback.URL)
	subscription.Topic, _ = url.Parse(topic.URL + "/")
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()

	ucase := usecase.NewSubscriptionUseCase(subscriptions, topics, callback.Client())

	ok, err := ucase.Subscribe(context.Background(), *subscription)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}

	if _, err := subscriptions.Get(context.Background(), subscription.SUID()); err != nil {
		t.Fatal(err)
	}

	t.Run("resubscribe", func(t *testing.T) {
		t.Parallel()

		ok, err := ucase.Subscribe(context.Background(), *subscription)
		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Errorf("want %t, got %t", true, ok)
		}
	})
}

func TestSubscriptionUseCase_Unsubscribe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
		fmt.Fprint(w, "hello, world")
	}))
	t.Cleanup(srv.Close)

	subscription := domain.TestSubscription(t, "https://example.com/")
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()

	if err := subscriptions.Create(context.Background(), subscription.SUID(), *subscription); err != nil {
		t.Fatal(err)
	}

	ok, err := usecase.NewSubscriptionUseCase(subscriptions, topics, srv.Client()).
		Unsubscribe(context.Background(), *subscription)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}

	if _, err := subscriptions.Get(context.Background(), subscription.SUID()); err == nil {
		t.Error("want error, got nil")
	}
}
