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
	"source.toby3d.me/toby3d/hub/internal/subscription"
)

type (
	Subscription struct {
		CreatedAt DateTime `db:"created_at"`
		UpdatedAt DateTime `db:"updated_at"`
		SyncedAt  DateTime `db:"synced_at"`
		DeleteAt  DateTime `db:"delete_at"`
		Topic     URL      `db:"topic"`
		Callback  URL      `db:"callback"`
		Secret    Secret   `db:"secret"`
	}

	DateTime struct {
		DateTime time.Time
		Valid    bool
	}

	URL struct {
		URL   *url.URL
		Valid bool
	}

	Secret struct {
		Secret domain.Secret
		Valid  bool
	}

	sqliteSubscriptionRepository struct {
		create *sqlx.NamedStmt
		update *sqlx.NamedStmt
		read   *sqlx.Stmt
		fetch  *sqlx.Stmt
		delete *sqlx.Stmt
	}
)

const (
	table      string = "subscriptions"
	queryTable string = `CREATE TABLE IF NOT EXISTS ` + table + ` (
		created_at DATETIME,
		updated_at DATETIME,
		synced_at DATETIME,
		delete_at DATETIME,
		topic TEXT,
		callback TEXT,
		secret TEXT,
		PRIMARY KEY (topic, callback)
	)`
	queryIndex  string = `CREATE INDEX idx_subscription ON ` + table + ` (topic, callback);`
	queryCreate string = `INSERT INTO ` + table + ` (created_at, updated_at, synced_at, delete_at, topic, ` +
		`callback, secret)
		VALUES (:created_at, :updated_at, :synced_at, :delete_at, :topic, :callback, :secret);`
	queryFetch  string = `SELECT * FROM ` + table + ` WHERE topic = ?;`
	queryRead   string = `SELECT * FROM ` + table + ` WHERE topic = ? AND callback = ?;`
	queryUpdate string = `UPDATE ` + table + `
				SET updated_at = :updated_at,
					synced_at = :synced_at,
					delete_at = :delete_at,
					secret = :secret
				WHERE topic = ? AND callback = ?;`
	queryDelete string = `DELETE FROM ` + table + ` WHERE topic = ? AND callback = ?;`
)

func NewSQLiteSubscriptionRepository(db *sqlx.DB) (subscription.Repository, error) {
	out := new(sqliteSubscriptionRepository)

	var err error
	if _, err = db.Exec(queryTable); err != nil {
		return nil, fmt.Errorf("subscription: sqlite: cannot prepare table: %w", err)
	}

	for q, dst := range map[string]**sqlx.NamedStmt{
		queryCreate: &out.create,
		queryUpdate: &out.update,
	} {
		if *dst, err = db.PrepareNamed(q); err != nil {
			return nil, fmt.Errorf("subscription: sqlite: cannot create prepared named subscription "+
				"statement: %w", err)
		}
	}

	for q, dst := range map[string]**sqlx.Stmt{
		queryDelete: &out.delete,
		queryFetch:  &out.fetch,
		queryRead:   &out.read,
	} {
		if *dst, err = db.Preparex(q); err != nil {
			return nil, fmt.Errorf("subscription: sqlite: cannot create prepared subscription statement: "+
				"%w", err)
		}
	}

	if _, err = db.Exec(queryIndex); err != nil {
		return nil, fmt.Errorf("subscription: sqlite: cannot create index: %w", err)
	}

	return out, nil
}

func (repo *sqliteSubscriptionRepository) Create(ctx context.Context, id domain.SUID, s domain.Subscription) error {
	row := new(Subscription)
	row.bind(s)

	if _, err := repo.create.ExecContext(ctx, row); err != nil {
		return fmt.Errorf("subscription: sqlite: cannot create subscription: %w", err)
	}

	return nil
}

func (repo *sqliteSubscriptionRepository) Get(ctx context.Context, id domain.SUID) (*domain.Subscription, error) {
	row := new(Subscription)
	if err := repo.read.GetContext(ctx, row, id.Topic().String(), id.Callback().String()); err != nil {
		return nil, fmt.Errorf("subscription: sqlite: cannot get subscription row: %w", err)
	}

	out := new(domain.Subscription)
	row.populate(out)

	return nil, nil
}

func (repo *sqliteSubscriptionRepository) Fetch(ctx context.Context, t *domain.Topic) ([]domain.Subscription, error) {
	rows, err := repo.fetch.QueryxContext(ctx, t.Self.String())
	if err != nil {
		return nil, fmt.Errorf("subscription: sqlite: cannot fetch subscription: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Subscription, 0)

	for rows.Next() {
		row := new(Subscription)
		if err = rows.StructScan(row); err != nil {
			return nil, fmt.Errorf("subscription: sqlite: cannot scan subscriptions row: %w", err)
		}

		var s domain.Subscription
		row.populate(&s)

		out = append(out, s)
	}

	return out, nil
}

func (repo *sqliteSubscriptionRepository) Update(ctx context.Context, id domain.SUID, update subscription.UpdateFunc) error {
	in, err := repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("subscription: sqlite: cannot find updating subscription: %w", err)
	}

	out, err := update(in)
	if err != nil {
		return fmt.Errorf("subscription: sqlite: cannot update subscription: %w", err)
	}

	row := new(Subscription)
	row.bind(*out)

	if _, err = repo.update.ExecContext(ctx, row); err != nil {
		return fmt.Errorf("subscription: sqlite: cannot update subscription row: %w", err)
	}

	return nil
}

func (repo *sqliteSubscriptionRepository) Delete(ctx context.Context, id domain.SUID) (bool, error) {
	result, err := repo.delete.ExecContext(ctx, id.Topic().String(), id.Callback().String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("subscription: sqlite: cannot delete subscription: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("subscription: sqlite: cannot read affected deleted rows result: %w", err)
	}

	return count == 1, nil
}

func (s *Subscription) bind(src domain.Subscription) {
	s.CreatedAt = NewDateTime(src.CreatedAt)
	s.UpdatedAt = NewDateTime(src.UpdatedAt)
	s.SyncedAt = NewDateTime(src.SyncedAt)
	s.DeleteAt = NewDateTime(src.ExpiredAt)
	s.Topic = NewURL(src.Topic)
	s.Callback = NewURL(src.Callback)
	s.Secret = NewSecret(src.Secret)
}

func (s Subscription) populate(dst *domain.Subscription) {
	dst.CreatedAt = s.CreatedAt.DateTime
	dst.UpdatedAt = s.UpdatedAt.DateTime
	dst.ExpiredAt = s.DeleteAt.DateTime
	dst.SyncedAt = s.SyncedAt.DateTime
	dst.Callback = s.Callback.URL
	dst.Topic = s.Topic.URL
	dst.Secret = s.Secret.Secret
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

func NewSecret(s domain.Secret) Secret {
	return Secret{
		Secret: s,
		Valid:  s.IsSet(),
	}
}

func (s *Secret) Scan(src any) error {
	var value string

	switch raw := src.(type) {
	default:
	case []byte:
		value = string(raw)
	case string:
		value = raw
	}

	parsed, err := domain.ParseSecret(value)
	if err != nil {
		return fmt.Errorf("Secret: cannot scan value as Secret: %w", err)
	}

	s.Secret = *parsed
	s.Valid = true

	return nil
}

func (s Secret) Value() (driver.Value, error) {
	if !s.Valid {
		return "", nil
	}

	return s.Secret.String(), nil
}
