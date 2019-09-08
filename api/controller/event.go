package controller

import (
	"encoding/json"

	"github.com/mylxsw/sync/collector"
)

type Event struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type EventPayload interface {
	Encode() []byte
}

func NewEvent(typ string, payload EventPayload) Event {
	return Event{
		Type:    typ,
		Payload: string(payload.Encode()),
	}
}

func (evt Event) Encode() []byte {
	rs, _ := json.Marshal(evt)
	return rs
}

type JobProgress struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Percentage float32 `json:"percentage"`
	Total      int     `json:"total"`
	Max        int     `json:"max"`
}

func (jrs JobProgress) Encode() []byte {
	rs, _ := json.Marshal(jrs)
	return rs
}

type JobRunningStatus struct {
	Console collector.StageMessage `json:"console"`
}

func (jrs JobRunningStatus) Encode() []byte {
	rs, _ := json.Marshal(jrs)
	return rs
}
