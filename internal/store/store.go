package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Incident struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Severity     string   `json:"severity"`
	Status       string   `json:"status"`
	Description  string   `json:"description"`
	Commander    string   `json:"commander"`
	CreatedAt    string   `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "inquest.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,\n\t\t\ttitle TEXT DEFAULT '',\n\t\t\tseverity TEXT DEFAULT 'medium',\n\t\t\tstatus TEXT DEFAULT 'investigating',\n\t\t\tdescription TEXT DEFAULT '',\n\t\t\tcommander TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (d *DB) Create(e *Incident) error {
	e.ID = genID()
	e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := d.db.Exec(`INSERT INTO incidents (id, title, severity, status, description, commander, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Title, e.Severity, e.Status, e.Description, e.Commander, e.CreatedAt)
	return err
}

func (d *DB) Get(id string) *Incident {
	row := d.db.QueryRow(`SELECT id, title, severity, status, description, commander, created_at FROM incidents WHERE id=?`, id)
	var e Incident
	if err := row.Scan(&e.ID, &e.Title, &e.Severity, &e.Status, &e.Description, &e.Commander, &e.CreatedAt); err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Incident {
	rows, err := d.db.Query(`SELECT id, title, severity, status, description, commander, created_at FROM incidents ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Incident
	for rows.Next() {
		var e Incident
		if err := rows.Scan(&e.ID, &e.Title, &e.Severity, &e.Status, &e.Description, &e.Commander, &e.CreatedAt); err != nil {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM incidents WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM incidents`).Scan(&n)
	return n
}
