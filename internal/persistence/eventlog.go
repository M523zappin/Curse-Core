package persistence

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
)

func chainHash(prevHash string, seq int64, prevState statemachine.State, evt statemachine.Event, nextState statemachine.State) string {
	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte(fmt.Sprintf("%d%d%d%d", seq, prevState, evt, nextState)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

type Entry struct {
	ID        string              `json:"id"`
	Sequence  int64               `json:"sequence"`
	PrevState statemachine.State  `json:"prev_state"`
	Event     statemachine.Event  `json:"event"`
	NewState  statemachine.State  `json:"new_state"`
	Timestamp time.Time           `json:"timestamp"`
	Data      json.RawMessage     `json:"data,omitempty"`
	Checksum  string              `json:"checksum"`
}

type EventLog struct {
	entries  []Entry
	sequence int64
	prevHash string
	filePath string
}

func NewEventLog(filePath string) *EventLog {
	return &EventLog{
		filePath: filePath,
		entries:  make([]Entry, 0),
	}
}

func (el *EventLog) Append(prev statemachine.State, event statemachine.Event, next statemachine.State, data interface{}) (Entry, error) {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return Entry{}, fmt.Errorf("marshal event data: %w", err)
		}
		raw = json.RawMessage(b)
	}
	el.sequence++
	entry := Entry{
		ID:        uuid.NewString(),
		Sequence:  el.sequence,
		PrevState: prev,
		Event:     event,
		NewState:  next,
		Timestamp: time.Now().UTC(),
		Data:      raw,
	}
	entry.Checksum = chainHash(el.prevHash, entry.Sequence, prev, event, next)
	el.prevHash = entry.Checksum
	el.entries = append(el.entries, entry)
	return entry, nil
}

func (el *EventLog) Entries() []Entry {
	return el.entries
}

func (el *EventLog) Sequence() int64 {
	return el.sequence
}

func (el *EventLog) LastHash() string {
	return el.prevHash
}

func (el *EventLog) Flush() error {
	if el.filePath == "" {
		return nil
	}
	var sb []byte
	for _, e := range el.entries {
		data, _ := json.Marshal(e)
		sb = append(sb, data...)
		sb = append(sb, '\n')
	}
	return writeFile(el.filePath, sb)
}

func LoadEventLog(filePath string) (*EventLog, error) {
	el := NewEventLog(filePath)
	data, err := readFile(filePath)
	if err != nil {
		return nil, err
	}
	lines := splitLines(string(data))
	prevHash := ""
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse entry: %w", err)
		}
		expected := chainHash(prevHash, entry.Sequence, entry.PrevState, entry.Event, entry.NewState)
		if entry.Checksum != expected {
			return nil, fmt.Errorf("chain integrity violation at seq %d: expected %s got %s",
				entry.Sequence, expected, entry.Checksum)
		}
		prevHash = entry.Checksum
		el.entries = append(el.entries, entry)
		el.sequence = entry.Sequence
		el.prevHash = entry.Checksum
	}
	return el, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
