package sqlite_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"source.toby3d.me/toby3d/hub/internal/domain"
	repository "source.toby3d.me/toby3d/hub/internal/topic/repository/sqlite"
)

// TODO(toby3d): All tests must be single purpose and parallel.
func Test(t *testing.T) {
	t.Parallel()

	tdb := sqlx.MustOpen("sqlite", filepath.Join(t.TempDir(), "testing.db"))
	t.Cleanup(func() { _ = tdb.Close() })

	repo, err := repository.NewSQLiteTopicRepository(tdb)
	if err != nil {
		t.Fatal(err)
	}

	topic := domain.TestTopic(t)

	// NOTE(toby3d): Create test.
	if err = repo.Create(context.Background(), topic.Self, *topic); err != nil {
		t.Fatal(err)
	}

	// NOTE(toby3d): Get test depends from Create.
	actual, err := repo.Get(context.Background(), topic.Self)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(actual, topic); diff != "" {
		t.Error(diff)
	}

	// NOTE(toby3d): Update test depend from Create.
	now := time.Now().UTC().Round(time.Second)
	content := []byte("lorem ipsum")

	if err = repo.Update(context.Background(), topic.Self, func(tx *domain.Topic) (*domain.Topic, error) {
		tx.Content = content
		tx.UpdatedAt = now

		return tx, nil
	}); err != nil {
		t.Fatal(err)
	}

	if actual, err = repo.Get(context.Background(), topic.Self); err != nil {
		t.Fatal(err)
	}

	if !actual.UpdatedAt.Equal(now) {
		t.Errorf("want '%s', got '%s'", now.Format(time.RFC3339), actual.UpdatedAt.Format(time.RFC3339))
	}

	if !bytes.Equal(actual.Content, content) {
		t.Errorf("want '%s', got '%s'", string(content), string(actual.Content))
	}

	/* NOTE(toby3d): Delete test depend from Create.
	ok, err := repo.Delete(context.Background(), topic.Self)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Errorf("want %t, got %t", true, ok)
	}

	if _, err = repo.Get(context.Background(), topic.Self); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("want %v error, got %v", sql.ErrNoRows, err)
	}
	*/
}
