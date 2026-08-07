// Package events is the wire contract for the Librescoot event bus.
//
// It is imported both by event-service, which derives events from existing
// Redis traffic, and by services that emit their own events directly. Keeping
// the struct, the topic names and the trigger vocabulary in one place is what
// stops the two from drifting.
package events

// Event is one occurrence on the bus.
//
// ID is assigned by Redis when the event is appended to the stream, so it is
// empty until then. From and To are promoted out of Data because "changed from
// X to Y" is the shape most rules match on.
type Event struct {
	ID    string         `json:"id,omitempty"`
	TS    int64          `json:"ts"`
	Topic string         `json:"topic"`
	Src   string         `json:"src"`
	From  string         `json:"from,omitempty"`
	To    string         `json:"to,omitempty"`
	Data  map[string]any `json:"data,omitempty"`
}

// New returns an Event with topic and src set. The caller fills in the rest.
// TS is left zero so callers can stamp it from their own clock, which keeps
// this package free of time dependencies and makes tests deterministic.
func New(topic, src string) Event {
	return Event{Topic: topic, Src: src}
}
