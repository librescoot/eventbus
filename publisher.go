package eventbus

import (
	"encoding/json"
	"fmt"
	"time"

	ipc "github.com/librescoot/redis-ipc"
)

// StreamKey is the Redis stream every event is appended to. Redis is not
// persisted on the scooter, so this is a live window for consumers that
// restart, not an archive.
const StreamKey = "events"

// ChannelPrefix is prepended to the topic to form the pub/sub channel name.
// Keeping events under their own prefix means PSUBSCRIBE ev:* cannot collide
// with the hash-notification channels that services already use.
const ChannelPrefix = "ev:"

const defaultStreamMaxLen = 2000

// Publisher writes events to the bus: one append to the stream and one
// publish on the topic channel per event.
type Publisher struct {
	client    *ipc.Client
	stream    *ipc.StreamPublisher
	src       string
	nowMillis func() int64
}

// PublisherOption configures a Publisher.
type PublisherOption func(*Publisher)

// WithStreamMaxLen overrides the approximate stream cap. Zero means the
// default of 2000.
func WithStreamMaxLen(n int64) PublisherOption {
	return func(p *Publisher) {
		if n > 0 {
			p.stream = p.client.NewStreamPublisher(StreamKey, ipc.WithMaxLen(n))
		}
	}
}

// WithClock replaces the millisecond clock. Tests use this to get
// deterministic timestamps.
func WithClock(fn func() int64) PublisherOption {
	return func(p *Publisher) { p.nowMillis = fn }
}

// NewPublisher returns a Publisher that stamps every event with src.
func NewPublisher(c *ipc.Client, src string, opts ...PublisherOption) *Publisher {
	p := &Publisher{
		client: c,
		stream: c.NewStreamPublisher(StreamKey, ipc.WithMaxLen(defaultStreamMaxLen)),
		src:    src,
		nowMillis: func() int64 {
			return time.Now().UnixMilli()
		},
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Emit writes e to the stream and publishes it on its topic channel.
//
// The stream ID is not known until the append succeeds, so the channel payload
// carries it: this is why the two writes cannot simply be pipelined blind.
//
// If the stream append succeeds and the publish then fails, the event is
// durably in the stream but never reaches live subscribers. Emit returns an
// error for that case same as any other, and a caller that retries will
// append a second copy of the event to the stream rather than lose it. The
// stream is the recovery path: a consumer that missed the publish can still
// pick the event up from there.
func (p *Publisher) Emit(e Event) error {
	e = p.stamp(e)

	streamBody, err := encodeJSON(e)
	if err != nil {
		return fmt.Errorf("encode event %s: %w", e.Topic, err)
	}

	id, err := p.stream.Add(map[string]any{"topic": e.Topic, "e": streamBody})
	if err != nil {
		return fmt.Errorf("append event to stream: %w", err)
	}
	e.ID = id

	channelBody, err := encodeJSON(e)
	if err != nil {
		return fmt.Errorf("encode event %s: %w", e.Topic, err)
	}

	if _, err := p.client.Publish(ChannelPrefix+e.Topic, channelBody); err != nil {
		return fmt.Errorf("publish event on %s: %w", ChannelPrefix+e.Topic, err)
	}
	return nil
}

func (p *Publisher) stamp(e Event) Event {
	if e.Src == "" {
		e.Src = p.src
	}
	if e.TS == 0 {
		e.TS = p.nowMillis()
	}
	return e
}

// encodeJSON marshals e. Data is a public map[string]any, so a producer can
// put something unmarshalable in it (a channel, a func, NaN); that is the
// caller's mistake, not this package's to crash the process over, so Emit
// reports it as an error instead of panicking.
func encodeJSON(e Event) (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
