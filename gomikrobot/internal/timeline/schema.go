package timeline

import (
	"time"
)

// TimelineEvent represents a single interaction in the history.
type TimelineEvent struct {
	ID             int64     `json:"id"`
	EventID        string    `json:"event_id"`       // Unique ID (e.g. WhatsApp MessageID)
	Timestamp      time.Time `json:"timestamp"`      // When it happened
	SenderID       string    `json:"sender_id"`      // Phone number
	SenderName     string    `json:"sender_name"`    // Display name
	EventType      string    `json:"event_type"`     // TEXT, AUDIO, IMAGE, SYSTEM
	ContentText    string    `json:"content_text"`   // The text or transcript
	MediaPath      string    `json:"media_path"`     // Path to local file if any
	VectorID       string    `json:"vector_id"`      // Qdrant ID
	Classification string    `json:"classification"` // ABM1 Category
	Authorized     bool      `json:"authorized"`     // Whether sender is in AllowFrom list
}

const Schema = `
CREATE TABLE IF NOT EXISTS timeline (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	event_id TEXT UNIQUE,
	timestamp DATETIME,
	sender_id TEXT,
	sender_name TEXT,
	event_type TEXT,
	content_text TEXT,
	media_path TEXT,
	vector_id TEXT,
	classification TEXT,
	authorized BOOLEAN DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_timeline_timestamp ON timeline(timestamp);
CREATE INDEX IF NOT EXISTS idx_timeline_sender ON timeline(sender_id);
CREATE INDEX IF NOT EXISTS idx_timeline_authorized ON timeline(authorized);

CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT,
	updated_at DATETIME
);
`
