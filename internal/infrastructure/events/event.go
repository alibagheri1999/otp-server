package events

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Version   string                 `json:"version"`
}

func NewEvent(eventType string, payload map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
		Source:    "otp-server",
		Version:   "1.0",
	}
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Event) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}
