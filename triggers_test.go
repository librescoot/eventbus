package events

import (
	"strings"
	"testing"
)

func TestValidTriggerAcceptsVocabulary(t *testing.T) {
	for _, s := range AllTriggers() {
		if !ValidTrigger(s) {
			t.Errorf("ValidTrigger(%q) = false, want true", s)
		}
	}
}

func TestValidTriggerRejectsUnknown(t *testing.T) {
	for _, s := range []string{"", "keycard", "auth", "auth.nope", "nope.keycard", "auth.keycard.extra"} {
		if ValidTrigger(s) {
			t.Errorf("ValidTrigger(%q) = true, want false", s)
		}
	}
}

func TestEveryTriggerHasAKnownPrefix(t *testing.T) {
	prefixes := map[string]bool{
		"rider": true, "auth": true, "command": true,
		"timer": true, "sensor": true, "system": true,
	}
	for _, s := range AllTriggers() {
		p, _, ok := strings.Cut(s, ".")
		if !ok {
			t.Errorf("trigger %q has no prefix", s)
			continue
		}
		if !prefixes[p] {
			t.Errorf("trigger %q has unknown prefix %q", s, p)
		}
	}
}

func TestAllTriggersHasNoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range AllTriggers() {
		if seen[s] {
			t.Errorf("duplicate trigger %q", s)
		}
		seen[s] = true
	}
}

func TestButtonTopic(t *testing.T) {
	cases := map[string]string{
		"horn|tap":        "button.horn.tap",
		"brake:left|hold": "button.brake.left.hold",
		"seatbox|press":   "button.seatbox.press",
	}
	for in, want := range cases {
		src, gesture, _ := strings.Cut(in, "|")
		if got := ButtonTopic(src, gesture); got != want {
			t.Errorf("ButtonTopic(%q, %q) = %q, want %q", src, gesture, got, want)
		}
	}
}
