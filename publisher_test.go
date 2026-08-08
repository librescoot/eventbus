package eventbus

import (
	"encoding/json"
	"testing"
	"time"

	ipc "github.com/librescoot/redis-ipc"
)

func testClient(t *testing.T) *ipc.Client {
	t.Helper()
	c, err := ipc.New(ipc.WithAddress("localhost"), ipc.WithPort(6379),
		ipc.WithDialTimeout(500*time.Millisecond))
	if err != nil {
		t.Skipf("no datastore on localhost:6379: %v", err)
	}
	if err := c.Ping(); err != nil {
		_ = c.Close()
		t.Skipf("datastore not responding: %v", err)
	}
	return c
}

func TestEmitPublishesOnTopicChannel(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	sub := c.Raw().Subscribe(t.Context(), "ev:battery.inserted")
	defer sub.Close()
	if _, err := sub.Receive(t.Context()); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	p := NewPublisher(c, "adapter", WithClock(func() int64 { return 42 }))
	e := New(TopicBatteryInserted, "adapter")
	e.To = "present"
	e.Data = map[string]any{"slot": 0}
	if err := p.Emit(e); err != nil {
		t.Fatalf("Emit: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		var got Event
		if err := json.Unmarshal([]byte(msg.Payload), &got); err != nil {
			t.Fatalf("payload not valid Event JSON: %v (%s)", err, msg.Payload)
		}
		if got.Topic != TopicBatteryInserted {
			t.Errorf("Topic = %q", got.Topic)
		}
		if got.TS != 42 {
			t.Errorf("TS = %d, want 42 (clock injection failed)", got.TS)
		}
		if got.ID == "" {
			t.Error("ID is empty, want the stream ID")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message on ev:battery.inserted within 2s")
	}
}

func TestEmitAppendsToStream(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	p := NewPublisher(c, "adapter", WithClock(func() int64 { return 7 }))
	e := New(TopicAlarmTriggered, "adapter")
	e.To = "level-2-triggered"
	if err := p.Emit(e); err != nil {
		t.Fatalf("Emit: %v", err)
	}

	res, err := c.Raw().XRevRangeN(t.Context(), StreamKey, "+", "-", 1).Result()
	if err != nil {
		t.Fatalf("XRevRangeN: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("got %d stream entries, want 1", len(res))
	}
	if res[0].Values["topic"] != TopicAlarmTriggered {
		t.Errorf("stream topic field = %v", res[0].Values["topic"])
	}
	raw, ok := res[0].Values["e"].(string)
	if !ok {
		t.Fatalf("stream entry has no string 'e' field: %v", res[0].Values)
	}
	var got Event
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("'e' field not valid Event JSON: %v", err)
	}
	if got.To != "level-2-triggered" {
		t.Errorf("To = %q", got.To)
	}
}

func TestEmitChannelAndStreamAgreeOnSameEvent(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	sub := c.Raw().Subscribe(t.Context(), "ev:battery.inserted")
	defer sub.Close()
	if _, err := sub.Receive(t.Context()); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	p := NewPublisher(c, "adapter", WithClock(func() int64 { return 99 }))
	e := New(TopicBatteryInserted, "adapter")
	e.To = "present"
	if err := p.Emit(e); err != nil {
		t.Fatalf("Emit: %v", err)
	}

	var fromChannel Event
	select {
	case msg := <-sub.Channel():
		if err := json.Unmarshal([]byte(msg.Payload), &fromChannel); err != nil {
			t.Fatalf("payload not valid Event JSON: %v (%s)", err, msg.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message on ev:battery.inserted within 2s")
	}

	res, err := c.Raw().XRevRangeN(t.Context(), StreamKey, "+", "-", 1).Result()
	if err != nil {
		t.Fatalf("XRevRangeN: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("got %d stream entries, want 1", len(res))
	}
	raw, ok := res[0].Values["e"].(string)
	if !ok {
		t.Fatalf("stream entry has no string 'e' field: %v", res[0].Values)
	}
	var fromStream Event
	if err := json.Unmarshal([]byte(raw), &fromStream); err != nil {
		t.Fatalf("'e' field not valid Event JSON: %v", err)
	}

	if fromChannel.ID == "" {
		t.Fatal("channel payload ID is empty")
	}
	if fromChannel.ID != res[0].ID {
		t.Errorf("channel ID = %q, stream entry ID = %q, want equal", fromChannel.ID, res[0].ID)
	}
	if fromChannel.Topic != fromStream.Topic {
		t.Errorf("channel Topic = %q, stream Topic = %q, want equal", fromChannel.Topic, fromStream.Topic)
	}
	if fromChannel.To != fromStream.To {
		t.Errorf("channel To = %q, stream To = %q, want equal", fromChannel.To, fromStream.To)
	}
}
