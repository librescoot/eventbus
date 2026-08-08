package eventbus

import "strings"

// The trigger vocabulary. data["trigger"] says why an event happened, in terms
// that mean the same thing whichever service emitted it. Service internals
// (FSM event IDs and the like) never reach the bus; producers map into this
// vocabulary instead. See EVENT-SERVICE-DESIGN.md section 6.4.
//
// The field is optional. A producer with no meaningful cause omits it rather
// than inventing one.
const (
	// rider.*: the rider physically did something.
	TriggerRiderKickstand = "rider.kickstand"
	TriggerRiderBrake     = "rider.brake"
	TriggerRiderSeatbox   = "rider.seatbox"
	TriggerRiderHorn      = "rider.horn"
	TriggerRiderBlinker   = "rider.blinker"
	TriggerRiderThrottle  = "rider.throttle"

	// auth.*: an identity was presented.
	TriggerAuthKeycard = "auth.keycard"
	TriggerAuthBLE     = "auth.ble"

	// command.*: something asked for this explicitly. The command queues carry
	// no provenance, so these say what was asked for, never who asked.
	TriggerCommandUnlock      = "command.unlock"
	TriggerCommandLock        = "command.lock"
	TriggerCommandHibernate   = "command.hibernate"
	TriggerCommandForceLock   = "command.force-lock"
	TriggerCommandHopOn       = "command.hop-on"
	TriggerCommandOpenSeatbox = "command.open-seatbox"

	// timer.*: a timer expired.
	TriggerTimerAutoStandby = "timer.auto-standby"
	TriggerTimerHibernation = "timer.hibernation"
	TriggerTimerShutdown    = "timer.shutdown"
	TriggerTimerSeatbox     = "timer.seatbox"

	// sensor.*: a measured condition was met.
	TriggerSensorMotion      = "sensor.motion"
	TriggerSensorVoltage     = "sensor.voltage"
	TriggerSensorTemperature = "sensor.temperature"
	TriggerSensorCharge      = "sensor.charge"

	// system.*: system lifecycle.
	TriggerSystemBoot      = "system.boot"
	TriggerSystemShutdown  = "system.shutdown"
	TriggerSystemUpdate    = "system.update"
	TriggerSystemDashboard = "system.dashboard"
	TriggerSystemFault     = "system.fault"
)

var allTriggers = []string{
	TriggerRiderKickstand, TriggerRiderBrake, TriggerRiderSeatbox,
	TriggerRiderHorn, TriggerRiderBlinker, TriggerRiderThrottle,
	TriggerAuthKeycard, TriggerAuthBLE,
	TriggerCommandUnlock, TriggerCommandLock, TriggerCommandHibernate,
	TriggerCommandForceLock, TriggerCommandHopOn, TriggerCommandOpenSeatbox,
	TriggerTimerAutoStandby, TriggerTimerHibernation, TriggerTimerShutdown,
	TriggerTimerSeatbox,
	TriggerSensorMotion, TriggerSensorVoltage, TriggerSensorTemperature,
	TriggerSensorCharge,
	TriggerSystemBoot, TriggerSystemShutdown, TriggerSystemUpdate,
	TriggerSystemDashboard, TriggerSystemFault,
}

var triggerSet = func() map[string]bool {
	m := make(map[string]bool, len(allTriggers))
	for _, s := range allTriggers {
		m[s] = true
	}
	return m
}()

// AllTriggers returns a copy of the vocabulary, for tests and documentation
// generation. Callers must not rely on ordering.
func AllTriggers() []string {
	out := make([]string, len(allTriggers))
	copy(out, allTriggers)
	return out
}

// ValidTrigger reports whether s is in the vocabulary. Producers should call
// this in tests, not at runtime: an invalid trigger is a programming error,
// and dropping the field at runtime would hide it.
func ValidTrigger(s string) bool {
	if strings.Count(s, ".") != 1 {
		return false
	}
	return triggerSet[s]
}
