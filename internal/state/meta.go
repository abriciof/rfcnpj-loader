package state

import (
	"context"
	"database/sql"
)

type MetaStore struct {
	db *sql.DB
}

func NewMetaStore(db *sql.DB) *MetaStore { return &MetaStore{db: db} }

func (m *MetaStore) Ensure(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS rfcnpj_meta (
  key text PRIMARY KEY,
  value text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);`)
	return err
}

func (m *MetaStore) Get(ctx context.Context, key string) (string, bool, error) {
	var v string
	err := m.db.QueryRowContext(ctx, `SELECT value FROM rfcnpj_meta WHERE key=$1`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (m *MetaStore) Set(ctx context.Context, key, value string) error {
	_, err := m.db.ExecContext(ctx, `
INSERT INTO rfcnpj_meta(key,value) VALUES ($1,$2)
ON CONFLICT (key) DO UPDATE SET value=excluded.value, updated_at=now()
`, key, value)
	return err
}
