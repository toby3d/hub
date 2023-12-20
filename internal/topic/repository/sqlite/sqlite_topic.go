package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"

	"source.toby3d.me/toby3d/hub/internal/domain"
	"source.toby3d.me/toby3d/hub/internal/topic"
)

type (
	Topic struct {
		CreatedAt   DateTime `db:"created_at"`
		UpdatedAt   DateTime `db:"updated_at"`
		URL         URL      `db:"url"`
		ContentType string   `db:"content_type"`
		Content     []byte   `db:"content"`
	}

	DateTime struct {
		DateTime time.Time
		Valid    bool
	}

	URL struct {
		URL   *url.URL
		Valid bool
	}

	sqliteTopicRepository struct {
		create *sqlx.NamedStmt
		update *sqlx.NamedStmt
		read   *sqlx.Stmt
		fetch  *sqlx.Stmt
		delete *sqlx.Stmt
	}
)

const (
	table      string = "topics"
	queryTable string = `CREATE TABLE IF NOT EXISTS ` + table + ` (
		created_at DATETIME,
		updated_at DATETIME,
		url TEXT PRIMARY KEY,
		content_type TEXT,
		content BLOB
	)`
	queryIndex  string = `CREATE INDEX IF NOT EXISTS idx_topic ON ` + table + ` (url);`
	queryCreate string = `INSERT INTO ` + table + ` (created_at, updated_at, url, content_type, content)
		       		VALUES (:created_at, :updated_at, :url, :content_type, :content);`
	queryFetch  string = `SELECT * FROM ` + table + `;`
	queryRead   string = `SELECT * FROM ` + table + ` WHERE url = ?;`
	queryUpdate string = `UPDATE ` + table + `
				SET updated_at = :updated_at,
					content_type = :content_type,
					content = :content
				WHERE url = :url;`
	queryDelete string = `DELETE FROM ` + table + ` WHERE url = ?;`
)

func NewSQLiteTopicRepository(db *sqlx.DB) (topic.Repository, error) {
	out := new(sqliteTopicRepository)

	var err error
	if _, err = db.Exec(queryTable); err != nil {
		return nil, fmt.Errorf("topic: sqlite: cannot prepare table: %w", err)
	}

	for q, dst := range map[string]**sqlx.NamedStmt{
		queryCreate: &out.create,
		queryUpdate: &out.update,
	} {
		if *dst, err = db.PrepareNamed(q); err != nil {
			return nil, fmt.Errorf("topic: sqlite: cannot create prepared named topic statement: %w", err)
		}
	}

	for q, dst := range map[string]**sqlx.Stmt{
		queryDelete: &out.delete,
		queryFetch:  &out.fetch,
		queryRead:   &out.read,
	} {
		if *dst, err = db.Preparex(q); err != nil {
			return nil, fmt.Errorf("topic: sqlite: cannot create prepared topic statement: %w", err)
		}
	}

	if _, err = db.Exec(queryIndex); err != nil {
		return nil, fmt.Errorf("topic: sqlite: cannot create index: %w", err)
	}

	return out, nil
}

func (repo *sqliteTopicRepository) Create(ctx context.Context, u *url.URL, t domain.Topic) error {
	row := new(Topic)
	row.bind(t)

	if _, err := repo.create.ExecContext(ctx, row); err != nil {
		return fmt.Errorf("topic: sqlite: cannot create topic: %w", err)
	}

	return nil
}

func (repo *sqliteTopicRepository) Fetch(ctx context.Context) ([]domain.Topic, error) {
	rows, err := repo.fetch.QueryxContext(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("topic: sqlite: cannot fetch topics: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Topic, 0)

	for rows.Next() {
		row := new(Topic)
		if err = rows.StructScan(row); err != nil {
			return nil, fmt.Errorf("topic: sqlite: cannot scan topics row: %w", err)
		}

		var t domain.Topic
		row.populate(&t)

		out = append(out, t)
	}

	return out, nil
}

func (repo *sqliteTopicRepository) Get(ctx context.Context, u *url.URL) (*domain.Topic, error) {
	row := new(Topic)
	if err := repo.read.GetContext(ctx, row, u.String()); err != nil {
		return nil, fmt.Errorf("topic: sqlite: cannot get topic row: %w", err)
	}

	out := new(domain.Topic)
	row.populate(out)

	return out, nil
}

func (repo *sqliteTopicRepository) Update(ctx context.Context, u *url.URL, update topic.UpdateFunc) error {
	in, err := repo.Get(ctx, u)
	if err != nil {
		return fmt.Errorf("topic: sqlite: cannot find updating topic: %w", err)
	}

	out, err := update(in)
	if err != nil {
		return fmt.Errorf("topic: sqlite: cannot update topic: %w", err)
	}

	row := new(Topic)
	row.bind(*out)

	if _, err = repo.update.ExecContext(ctx, row); err != nil {
		return fmt.Errorf("topic: sqlite: cannot update topic row: %w", err)
	}

	return nil
}

func (repo *sqliteTopicRepository) Delete(ctx context.Context, u *url.URL) (bool, error) {
	result, err := repo.delete.ExecContext(ctx, u.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("topic: sqlite: cannot delete topic: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("topic: sqlite: cannot read affected deleted rows result: %w", err)
	}

	return count == 1, nil
}

func (t *Topic) bind(src domain.Topic) {
	t.Content = src.Content
	t.ContentType = src.ContentType
	t.CreatedAt = NewDateTime(src.CreatedAt)
	t.UpdatedAt = NewDateTime(src.UpdatedAt)
	t.URL = NewURL(src.Self)
}

func (t Topic) populate(dst *domain.Topic) {
	dst.Content = t.Content
	dst.ContentType = t.ContentType
	dst.CreatedAt = t.CreatedAt.DateTime
	dst.Self = t.URL.URL
	dst.UpdatedAt = t.UpdatedAt.DateTime
}

func NewURL(u *url.URL) URL {
	return URL{
		URL:   u,
		Valid: u != nil,
	}
}

func (u *URL) Scan(src any) error {
	var err error

	switch s := src.(type) {
	case []byte:
		if u.URL, err = url.Parse(string(s)); err != nil {
			return fmt.Errorf("URL: cannot scan BLOB value as URL: %w", err)
		}

		u.Valid = true
	case string:
		if u.URL, err = url.Parse(s); err != nil {
			return fmt.Errorf("URL: cannot scan TEXT value as URL: %w", err)
		}

		u.Valid = true
	}

	return nil
}

func (u URL) Value() (driver.Value, error) {
	if !u.Valid {
		return "", nil
	}

	return u.URL.String(), nil
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{
		DateTime: t,
		Valid:    !t.IsZero(),
	}
}

func (dt *DateTime) Scan(src any) error {
	switch s := src.(type) {
	case int64:
		dt.DateTime = time.Unix(s, 0)
		dt.Valid = true
	}

	return nil
}

func (dt DateTime) Value() (driver.Value, error) {
	if !dt.Valid {
		return 0, nil
	}

	return dt.DateTime.Unix(), nil
}
