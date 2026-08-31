package bootstrap

import (
	"context"
	"errors"
	"time"

	"github.com/go-rio/rio"
	"github.com/libtnb/sessions/driver"
)

// sessionRow maps to "sessions"; CreatedAt/UpdatedAt are stamped by rio.
type sessionRow struct {
	ID        string `rio:",pk"`
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (sessionRow) TableName() string { return "sessions" }

var (
	sessionByIDQuery     = rio.From[sessionRow]().Where("id = ?").Must()
	expiredSessionsQuery = rio.From[sessionRow]().Where("updated_at < ?").Must()
)

// driver.Driver is context-free, so each method runs on a background context.
type sessionStore struct {
	db *rio.DB
}

func newSessionStore(db *rio.DB) (driver.Driver, error) {
	_, err := rio.Exec(context.Background(), db, `CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL DEFAULT '',
		created_at DATETIME,
		updated_at DATETIME
	)`)
	if err != nil {
		return nil, err
	}
	return &sessionStore{db: db}, nil
}

func (s *sessionStore) Close() error { return nil }

func (s *sessionStore) Destroy(id string) error {
	_, err := sessionByIDQuery.DeleteAll(context.Background(), s.db, id)
	return err
}

func (s *sessionStore) Read(id string) (string, bool, error) {
	row, err := rio.Find[sessionRow](context.Background(), s.db, id)
	if err != nil {
		if errors.Is(err, rio.ErrNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return row.Data, true, nil
}

func (s *sessionStore) Touch(id string) (bool, error) {
	n, err := sessionByIDQuery.UpdateAll(
		context.Background(),
		s.db,
		rio.Set{"updated_at": time.Now()},
		id,
	)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *sessionStore) Gc(maxLifetime int) error {
	cutoff := time.Now().Add(-time.Duration(maxLifetime) * time.Second)
	_, err := expiredSessionsQuery.DeleteAll(context.Background(), s.db, cutoff)
	return err
}

func (s *sessionStore) Write(id string, data string) error {
	row := &sessionRow{ID: id, Data: data}
	return rio.Upsert(context.Background(), s.db, row, rio.OnConflict("id"), rio.DoUpdate("data"))
}
