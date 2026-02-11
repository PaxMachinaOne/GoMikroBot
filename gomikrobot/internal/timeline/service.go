package timeline

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type TimelineService struct {
	db *sql.DB
}

func NewTimelineService(dbPath string) (*TimelineService, error) {
	db, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("failed to open timeline db: %w", err)
	}

	// Apply schema
	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply schema: %w", err)
	}

	return &TimelineService{db: db}, nil
}

func (s *TimelineService) Close() error {
	return s.db.Close()
}

func (s *TimelineService) AddEvent(evt *TimelineEvent) error {
	query := `
	INSERT INTO timeline (event_id, timestamp, sender_id, sender_name, event_type, content_text, media_path, vector_id, classification, authorized)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		evt.EventID,
		evt.Timestamp,
		evt.SenderID,
		evt.SenderName,
		evt.EventType,
		evt.ContentText,
		evt.MediaPath,
		evt.VectorID,
		evt.Classification,
		evt.Authorized,
	)
	return err
}

type FilterArgs struct {
	SenderID       string
	Limit          int
	Offset         int
	StartDate      *time.Time
	EndDate        *time.Time
	AuthorizedOnly *bool // nil = all, true = authorized only, false = unauthorized only
}

func (s *TimelineService) GetEvents(filter FilterArgs) ([]TimelineEvent, error) {
	query := `SELECT id, event_id, timestamp, sender_id, sender_name, event_type, content_text, media_path, vector_id, classification, authorized FROM timeline WHERE 1=1`
	args := []interface{}{}

	if filter.SenderID != "" {
		query += " AND sender_id = ?"
		args = append(args, filter.SenderID)
	}
	if filter.StartDate != nil {
		query += " AND timestamp >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND timestamp <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.AuthorizedOnly != nil {
		query += " AND authorized = ?"
		args = append(args, *filter.AuthorizedOnly)
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []TimelineEvent
	for rows.Next() {
		var e TimelineEvent
		err := rows.Scan(
			&e.ID,
			&e.EventID,
			&e.Timestamp,
			&e.SenderID,
			&e.SenderName,
			&e.EventType,
			&e.ContentText,
			&e.MediaPath,
			&e.VectorID,
			&e.Classification,
			&e.Authorized,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// GetSetting returns a setting value by key.
func (s *TimelineService) GetSetting(key string) (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// SetSetting persists a setting value.
func (s *TimelineService) SetSetting(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value, updated_at) VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value)
	return err
}

// IsSilentMode checks if silent mode is enabled. Defaults to true (safe default).
func (s *TimelineService) IsSilentMode() bool {
	val, err := s.GetSetting("silent_mode")
	if err != nil {
		return true // Safe default: silent
	}
	return val == "true"
}
