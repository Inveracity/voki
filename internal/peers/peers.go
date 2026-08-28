package peers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
	Rejected Status = "rejected"
)

type Peer struct {
	ID          int       `json:"id"`
	Fingerprint string    `json:"fingerprint"`
	Name        string    `json:"name"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type DB struct {
	NextID int     `json:"next_id"`
	Peers  []Peer  `json:"peers"`

	path string `json:"-"`
}

// Load reads the peer database. A missing file gives an empty database.
func Load(path string) (*DB, error) {
	db := &DB{NextID: 1, path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return db, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, db); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	db.path = path
	return db, nil
}

// ponytail: whole-file rewrite per change, not process-safe under concurrent writers.
// Fine while mutations are rare (bond, accept, reject). Add file locking if bond storms happen.
func (db *DB) Save() error {
	if err := os.MkdirAll(filepath.Dir(db.path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	tmp := db.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, db.path)
}

func (db *DB) StatusOf(fingerprint string) Status {
	for _, p := range db.Peers {
		if p.Fingerprint == fingerprint {
			return p.Status
		}
	}
	return ""
}

// UpsertPending registers a fingerprint as pending and returns its id.
// An existing peer keeps its id and status.
func (db *DB) UpsertPending(fingerprint, name string) (int, Status) {
	for _, p := range db.Peers {
		if p.Fingerprint == fingerprint {
			return p.ID, p.Status
		}
	}
	id := db.NextID
	db.NextID++
	db.Peers = append(db.Peers, Peer{
		ID:          id,
		Fingerprint: fingerprint,
		Name:        name,
		Status:      Pending,
		CreatedAt:   time.Now().UTC(),
	})
	return id, Pending
}

func (db *DB) SetStatus(id int, status Status) error {
	for i := range db.Peers {
		if db.Peers[i].ID == id {
			db.Peers[i].Status = status
			return db.Save()
		}
	}
	return fmt.Errorf("no peer with id %d", id)
}
