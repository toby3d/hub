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
	topicmemoryrepo "source.toby3d.me/toby3d/hub/internal/topic/repository/memory"
	"source.toby3d.me/toby3d/hub/internal/topic/usecase"
)

func TestTopicUseCase_Publish(t *testing.T) {
	t.Parallel()

	topic := domain.TestTopic(t)
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(common.HeaderContentType, topic.ContentType)
		fmt.Fprint(w, topic.Content)
	}))
	t.Cleanup(srv.Close)

	topic.Self, _ = url.Parse(srv.URL + "/")

	ok, err := usecase.NewTopicUseCase(topics, srv.Client()).
		Publish(context.Background(), topic.Self)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}

	if _, err := topics.Get(context.Background(), topic.Self); err != nil {
		t.Fatal(err)
	}
}
