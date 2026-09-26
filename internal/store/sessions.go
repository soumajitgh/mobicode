package store

import (
	"database/sql"
	"errors"
	"time"
)

type SessionStore struct{ DB *sql.DB }

func (s SessionStore) Delete(token string) error {
	_, err := s.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (s SessionStore) Find(token string) ([]byte, bool, error) {
	var data []byte
	err := s.DB.QueryRow("SELECT data FROM sessions WHERE token = ? AND expiry > ?", token, time.Now().UTC()).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return data, err == nil, err
}

func (s SessionStore) Commit(token string, data []byte, expiry time.Time) error {
	_, err := s.DB.Exec("INSERT INTO sessions (token, data, expiry) VALUES (?, ?, ?) ON CONFLICT(token) DO UPDATE SET data=excluded.data, expiry=excluded.expiry", token, data, expiry.UTC())
	return err
}
