package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	_ "modernc.org/sqlite"
)

type BlobStore struct {
	db *sql.DB
}

func NewBlobStore(dbPath string) (*BlobStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS blobs (
		hash TEXT PRIMARY KEY,
		data BLOB NOT NULL
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &BlobStore{db: db}, nil
}

func (s *BlobStore) Put(data []byte) (string, error) {
	h := sha256.Sum256(data)
	hash := hex.EncodeToString(h[:])
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO blobs (hash, data) VALUES (?, ?)`,
		hash, data,
	)
	if err != nil {
		return "", fmt.Errorf("put blob: %w", err)
	}
	return hash, nil
}

func (s *BlobStore) Get(hash string) ([]byte, error) {
	var data []byte
	err := s.db.QueryRow(`SELECT data FROM blobs WHERE hash = ?`, hash).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get blob: %w", err)
	}
	return data, nil
}

func (s *BlobStore) Close() error {
	return s.db.Close()
}
