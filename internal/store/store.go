package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type Incident struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Severity    string   `json:"severity"` // sev1, sev2, sev3, sev4
	Status      string   `json:"status"`   // investigating, identified, monitoring, resolved
	Lead        string   `json:"lead,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Services    []string `json:"services"`
	StartedAt   string   `json:"started_at"`
	ResolvedAt  string   `json:"resolved_at,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdateCount int      `json:"update_count"`
	HasPostmortem bool   `json:"has_postmortem"`
}

type TimelineEntry struct {
	ID         string `json:"id"`
	IncidentID string `json:"incident_id"`
	Author     string `json:"author,omitempty"`
	Message    string `json:"message"`
	Status     string `json:"status,omitempty"` // status at time of update
	CreatedAt  string `json:"created_at"`
}

type Postmortem struct {
	ID          string `json:"id"`
	IncidentID  string `json:"incident_id"`
	WhatHappened string `json:"what_happened"`
	RootCause   string `json:"root_cause"`
	ActionItems string `json:"action_items"`
	Lessons     string `json:"lessons"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type IncidentFilter struct {
	Status   string
	Severity string
	Lead     string
	Search   string
	Limit    int
	Offset   int
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
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			severity TEXT DEFAULT 'sev3',
			status TEXT DEFAULT 'investigating',
			lead TEXT DEFAULT '',
			summary TEXT DEFAULT '',
			services_json TEXT DEFAULT '[]',
			started_at TEXT DEFAULT (datetime('now')),
			resolved_at TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS timeline (
			id TEXT PRIMARY KEY,
			incident_id TEXT NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
			author TEXT DEFAULT '',
			message TEXT NOT NULL,
			status TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS postmortems (
			id TEXT PRIMARY KEY,
			incident_id TEXT UNIQUE NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
			what_happened TEXT DEFAULT '',
			root_cause TEXT DEFAULT '',
			action_items TEXT DEFAULT '',
			lessons TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_timeline_incident ON timeline(incident_id)`,
		`CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }

func calcDuration(start, end string) string {
	s, _ := time.Parse(time.RFC3339, start)
	e, _ := time.Parse(time.RFC3339, end)
	if e.IsZero() {
		e = time.Now()
	}
	dur := e.Sub(s)
	if dur < time.Minute {
		return fmt.Sprintf("%ds", int(dur.Seconds()))
	} else if dur < time.Hour {
		return fmt.Sprintf("%dm", int(dur.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(dur.Hours()), int(dur.Minutes())%60)
}

// ── Incidents ──

func (d *DB) CreateIncident(inc *Incident) error {
	inc.ID = genID()
	inc.CreatedAt = now()
	if inc.StartedAt == "" {
		inc.StartedAt = inc.CreatedAt
	}
	if inc.Status == "" {
		inc.Status = "investigating"
	}
	if inc.Severity == "" {
		inc.Severity = "sev3"
	}
	if inc.Services == nil {
		inc.Services = []string{}
	}
	sj, _ := json.Marshal(inc.Services)
	_, err := d.db.Exec(`INSERT INTO incidents (id,title,severity,status,lead,summary,services_json,started_at,created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		inc.ID, inc.Title, inc.Severity, inc.Status, inc.Lead, inc.Summary, string(sj), inc.StartedAt, inc.CreatedAt)
	if err != nil {
		return err
	}
	// auto-add timeline entry
	d.addTimeline(inc.ID, inc.Lead, "Incident declared: "+inc.Title, inc.Status)
	return nil
}

func (d *DB) hydrateIncident(inc *Incident) {
	d.db.QueryRow(`SELECT COUNT(*) FROM timeline WHERE incident_id=?`, inc.ID).Scan(&inc.UpdateCount)
	var pmID string
	inc.HasPostmortem = d.db.QueryRow(`SELECT id FROM postmortems WHERE incident_id=?`, inc.ID).Scan(&pmID) == nil
	if inc.ResolvedAt != "" {
		inc.Duration = calcDuration(inc.StartedAt, inc.ResolvedAt)
	} else {
		inc.Duration = calcDuration(inc.StartedAt, "")
	}
}

func (d *DB) scanIncident(s interface{ Scan(...any) error }) *Incident {
	var inc Incident
	var sj string
	if err := s.Scan(&inc.ID, &inc.Title, &inc.Severity, &inc.Status, &inc.Lead, &inc.Summary, &sj, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt); err != nil {
		return nil
	}
	json.Unmarshal([]byte(sj), &inc.Services)
	if inc.Services == nil {
		inc.Services = []string{}
	}
	d.hydrateIncident(&inc)
	return &inc
}

const incCols = `id,title,severity,status,lead,summary,services_json,started_at,resolved_at,created_at`

func (d *DB) GetIncident(id string) *Incident {
	return d.scanIncident(d.db.QueryRow(`SELECT `+incCols+` FROM incidents WHERE id=?`, id))
}

func (d *DB) ListIncidents(f IncidentFilter) ([]Incident, int) {
	where := []string{"1=1"}
	args := []any{}
	if f.Status != "" && f.Status != "all" {
		if f.Status == "active" {
			where = append(where, "status != 'resolved'")
		} else {
			where = append(where, "status=?")
			args = append(args, f.Status)
		}
	}
	if f.Severity != "" {
		where = append(where, "severity=?")
		args = append(args, f.Severity)
	}
	if f.Lead != "" {
		where = append(where, "lead=?")
		args = append(args, f.Lead)
	}
	if f.Search != "" {
		where = append(where, "(title LIKE ? OR summary LIKE ?)")
		s := "%" + f.Search + "%"
		args = append(args, s, s)
	}
	w := strings.Join(where, " AND ")
	var total int
	d.db.QueryRow("SELECT COUNT(*) FROM incidents WHERE "+w, args...).Scan(&total)
	if f.Limit <= 0 {
		f.Limit = 50
	}
	q := fmt.Sprintf("SELECT %s FROM incidents WHERE %s ORDER BY CASE severity WHEN 'sev1' THEN 0 WHEN 'sev2' THEN 1 WHEN 'sev3' THEN 2 WHEN 'sev4' THEN 3 END, started_at DESC LIMIT ? OFFSET ?", incCols, w)
	args = append(args, f.Limit, f.Offset)
	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()
	var out []Incident
	for rows.Next() {
		if inc := d.scanIncident(rows); inc != nil {
			out = append(out, *inc)
		}
	}
	return out, total
}

func (d *DB) UpdateIncident(id string, inc *Incident) error {
	sj, _ := json.Marshal(inc.Services)
	_, err := d.db.Exec(`UPDATE incidents SET title=?,severity=?,status=?,lead=?,summary=?,services_json=? WHERE id=?`,
		inc.Title, inc.Severity, inc.Status, inc.Lead, inc.Summary, string(sj), id)
	return err
}

func (d *DB) UpdateStatus(id, status, author string) error {
	_, err := d.db.Exec(`UPDATE incidents SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	if status == "resolved" {
		d.db.Exec(`UPDATE incidents SET resolved_at=? WHERE id=?`, now(), id)
	}
	d.addTimeline(id, author, "Status changed to "+status, status)
	return nil
}

func (d *DB) DeleteIncident(id string) error {
	d.db.Exec(`DELETE FROM timeline WHERE incident_id=?`, id)
	d.db.Exec(`DELETE FROM postmortems WHERE incident_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM incidents WHERE id=?`, id)
	return err
}

// ── Timeline ──

func (d *DB) addTimeline(incidentID, author, message, status string) {
	d.db.Exec(`INSERT INTO timeline (id,incident_id,author,message,status,created_at) VALUES (?,?,?,?,?,?)`,
		genID(), incidentID, author, message, status, now())
}

func (d *DB) AddUpdate(incidentID, author, message string) error {
	inc := d.GetIncident(incidentID)
	if inc == nil {
		return fmt.Errorf("incident not found")
	}
	d.addTimeline(incidentID, author, message, inc.Status)
	return nil
}

func (d *DB) ListTimeline(incidentID string) []TimelineEntry {
	rows, err := d.db.Query(`SELECT id,incident_id,author,message,status,created_at FROM timeline WHERE incident_id=? ORDER BY created_at ASC`, incidentID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []TimelineEntry
	for rows.Next() {
		var t TimelineEntry
		if err := rows.Scan(&t.ID, &t.IncidentID, &t.Author, &t.Message, &t.Status, &t.CreatedAt); err != nil {
			continue
		}
		out = append(out, t)
	}
	return out
}

// ── Postmortems ──

func (d *DB) CreatePostmortem(pm *Postmortem) error {
	pm.ID = genID()
	pm.CreatedAt = now()
	pm.UpdatedAt = pm.CreatedAt
	_, err := d.db.Exec(`INSERT INTO postmortems (id,incident_id,what_happened,root_cause,action_items,lessons,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`,
		pm.ID, pm.IncidentID, pm.WhatHappened, pm.RootCause, pm.ActionItems, pm.Lessons, pm.CreatedAt, pm.UpdatedAt)
	return err
}

func (d *DB) GetPostmortem(incidentID string) *Postmortem {
	var pm Postmortem
	if err := d.db.QueryRow(`SELECT id,incident_id,what_happened,root_cause,action_items,lessons,created_at,updated_at FROM postmortems WHERE incident_id=?`, incidentID).Scan(&pm.ID, &pm.IncidentID, &pm.WhatHappened, &pm.RootCause, &pm.ActionItems, &pm.Lessons, &pm.CreatedAt, &pm.UpdatedAt); err != nil {
		return nil
	}
	return &pm
}

func (d *DB) UpdatePostmortem(incidentID string, pm *Postmortem) error {
	_, err := d.db.Exec(`UPDATE postmortems SET what_happened=?,root_cause=?,action_items=?,lessons=?,updated_at=? WHERE incident_id=?`,
		pm.WhatHappened, pm.RootCause, pm.ActionItems, pm.Lessons, now(), incidentID)
	return err
}

// ── Stats ──

type Stats struct {
	Total       int            `json:"total"`
	Active      int            `json:"active"`
	Resolved    int            `json:"resolved"`
	BySeverity  map[string]int `json:"by_severity"`
	MTTR        string         `json:"mttr"` // mean time to resolve
	Postmortems int            `json:"postmortems"`
}

func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM incidents`).Scan(&s.Total)
	d.db.QueryRow(`SELECT COUNT(*) FROM incidents WHERE status!='resolved'`).Scan(&s.Active)
	d.db.QueryRow(`SELECT COUNT(*) FROM incidents WHERE status='resolved'`).Scan(&s.Resolved)
	d.db.QueryRow(`SELECT COUNT(*) FROM postmortems`).Scan(&s.Postmortems)
	s.BySeverity = map[string]int{}
	rows, _ := d.db.Query(`SELECT severity, COUNT(*) FROM incidents GROUP BY severity`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var sv string
			var c int
			rows.Scan(&sv, &c)
			s.BySeverity[sv] = c
		}
	}
	// MTTR
	var totalMin float64
	var count int
	rrows, _ := d.db.Query(`SELECT started_at, resolved_at FROM incidents WHERE resolved_at != ''`)
	if rrows != nil {
		defer rrows.Close()
		for rrows.Next() {
			var sa, ra string
			rrows.Scan(&sa, &ra)
			start, _ := time.Parse(time.RFC3339, sa)
			end, _ := time.Parse(time.RFC3339, ra)
			totalMin += end.Sub(start).Minutes()
			count++
		}
	}
	if count > 0 {
		avg := totalMin / float64(count)
		if avg < 60 {
			s.MTTR = fmt.Sprintf("%.0fm", avg)
		} else {
			s.MTTR = fmt.Sprintf("%.1fh", avg/60)
		}
	} else {
		s.MTTR = "-"
	}
	return s
}
