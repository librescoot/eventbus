package eventbus

import (
	"encoding/json"
	"testing"
)

func TestEventJSONOmitsEmptyOptionalFields(t *testing.T) {
	e := New("battery.inserted", "adapter")
	e.TS = 1780000000123
	e.Data = map[string]any{"slot": 0}

	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	want := `{"ts":1780000000123,"topic":"battery.inserted","src":"adapter","data":{"slot":0}}`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestEventJSONIncludesFromTo(t *testing.T) {
	e := New("vehicle.unlocked", "adapter")
	e.TS = 1
	e.From = "stand-by"
	e.To = "parked"

	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"ts":1,"topic":"vehicle.unlocked","src":"adapter","from":"stand-by","to":"parked"}`
	if string(b) != want {
		t.Errorf("got  %s\nwant %s", b, want)
	}
}

func TestNewSetsTopicAndSrc(t *testing.T) {
	e := New("a.b", "svc")
	if e.Topic != "a.b" || e.Src != "svc" {
		t.Errorf("New() = %+v", e)
	}
}
