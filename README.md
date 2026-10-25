# Librescoot Event Bus

Wire contract for the Librescoot event bus: the event envelope, topic names,
trigger vocabulary, and publisher.

Part of the [Librescoot](https://librescoot.org/) open-source platform.

## Overview

`eventbus` is a small Go library imported by [event-service](https://github.com/librescoot/event-service),
which derives events from existing Redis traffic, and by services that emit
their own events directly. Keeping the envelope struct, the topic names, and
the trigger vocabulary in one place is what stops the two from drifting.

## Event envelope

One occurrence on the bus is a JSON `Event`:

| Field | Type | Meaning |
|-------|------|---------|
| `id` | string | Assigned by Redis when the event is appended to the stream; empty until then |
| `ts` | int64 | Unix milliseconds; the producer stamps it from its own clock |
| `topic` | string | Dotted topic name, see below |
| `src` | string | Emitting service |
| `from` / `to` | string | Promoted out of `data`; "changed from X to Y" is the shape most rules match on |
| `data` | map | Free-form payload; `trigger` carries the trigger vocabulary, `slot` the battery slot |

`eventbus.New(topic, src)` returns a skeleton event with topic and source set.

## Topics

Topic names are dotted so subscribers can filter server-side with
`PSUBSCRIBE ev:vehicle.*` rather than decoding every message. The battery slot
is carried in `data["slot"]`, not in the topic, so a glob over battery events
does not have to enumerate slots.

Named transitions cover vehicle lock/unlock/hibernating, ride started/ended,
seatbox, handlebar, kickstand, blinker, keycard authentication, battery
insertion/removal/state/charge, auxiliary and control-board battery charge,
power state and wake, alarm arming/disarm/trigger, motion, and ECU faults.
Alongside the named transitions, the complete-record topics
`vehicle.state.changed` and `alarm.status.changed` fire for every vehicle state
transition and alarm status change respectively.

Some topics are published vocabulary only in v1 (not yet wired to the
adapter): `ecu.fault.raised` / `ecu.fault.cleared` and `gps.*`. Deriving them
needs producer-side changes, not an adapter fix — see the design notes in the
repository.

## Triggers

`data["trigger"]` says why an event happened, in terms that mean the same thing
whichever service emitted it. Service internals (FSM event IDs and the like)
never reach the bus; producers map into this vocabulary instead. The field is
optional: a producer with no meaningful cause omits it rather than inventing
one.

| Prefix | Meaning | Examples |
|--------|---------|----------|
| `rider.*` | The rider physically did something | `rider.kickstand`, `rider.brake`, `rider.throttle` |
| `auth.*` | An identity was presented | `auth.keycard`, `auth.ble` |
| `command.*` | Something asked for this explicitly | `command.unlock`, `command.hibernate`, `command.open-seatbox` |
| `timer.*` | A timer expired | `timer.auto-standby`, `timer.hibernation` |
| `sensor.*` | A measured condition was met | `sensor.motion`, `sensor.voltage` |
| `system.*` | System lifecycle | `system.boot`, `system.update`, `system.fault` |

## Publishing

The `Publisher` writes each event twice: one append to the Redis stream
`events` (capped at roughly 2000 entries — Redis is not persisted on the
scooter, so the stream is a live window for consumers that restart, not an
archive) and one publish on the topic's channel, prefixed `ev:` so it can never
collide with the hash-notification channels services already use.

```go
import "github.com/librescoot/eventbus"

publisher := eventbus.NewPublisher(client, "vehicle-service")
event := eventbus.New(eventbus.TopicVehicleLocked, "vehicle-service")
publisher.Emit(event)
```

## Consuming

Read events either from the `events` stream (XREAD, keeps a replay window for
consumers that restart) or by subscribing to `ev:*` channels for live
notification.

## License

Original Librescoot code in this repository is available, at your option, under
[AGPL-3.0-only](LICENSE) or [CC-BY-NC-SA-4.0](LICENSE.CC-BY-NC-SA-4.0).
Third-party code retains its existing license.

Made with ❤️ by the Librescoot community