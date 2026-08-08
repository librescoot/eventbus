package eventbus

// Topic names. Dotted so subscribers can filter server-side with
// PSUBSCRIBE ev:vehicle.* rather than decoding every message.
//
// Battery slot is carried in data["slot"], not in the topic, so that a glob
// over battery events does not have to enumerate slots.
const (
	// Named vehicle transitions. Emitted alongside TopicVehicleStateChanged.
	TopicVehicleUnlocked    = "vehicle.unlocked"
	TopicVehicleLocked      = "vehicle.locked"
	TopicVehicleHibernating = "vehicle.hibernating"
	TopicRideStarted        = "ride.started"
	TopicRideEnded          = "ride.ended"

	// The complete record: fires for every vehicle state transition.
	TopicVehicleStateChanged = "vehicle.state.changed"

	TopicVehicleSeatboxOpened     = "vehicle.seatbox.opened"
	TopicVehicleSeatboxClosed     = "vehicle.seatbox.closed"
	TopicVehicleHandlebarLocked   = "vehicle.handlebar.locked"
	TopicVehicleHandlebarUnlocked = "vehicle.handlebar.unlocked"
	TopicVehicleKickstandUp       = "vehicle.kickstand.up"
	TopicVehicleKickstandDown     = "vehicle.kickstand.down"
	TopicVehicleBlinkerChanged    = "vehicle.blinker.changed"

	TopicKeycardAuthPassed = "keycard.auth.passed"
	TopicKeycardAuthFailed = "keycard.auth.failed"

	TopicBatteryInserted      = "battery.inserted"
	TopicBatteryRemoved       = "battery.removed"
	TopicBatteryStateChanged  = "battery.state.changed"
	TopicBatteryChargeChanged = "battery.charge.changed"
	TopicAuxChargeChanged     = "aux.charge.changed"
	TopicCBBChargeChanged     = "cbb.charge.changed"

	TopicPowerStateChanged = "power.state.changed"
	TopicPowerWake         = "power.wake"

	TopicAlarmArmed     = "alarm.armed"
	TopicAlarmDisarmed  = "alarm.disarmed"
	TopicAlarmTriggered = "alarm.triggered"

	TopicMotionDetected = "motion.detected"

	TopicECUFaultRaised  = "ecu.fault.raised"
	TopicECUFaultCleared = "ecu.fault.cleared"

	TopicGPSFixAcquired = "gps.fix.acquired"
	TopicGPSFixLost     = "gps.fix.lost"

	TopicNetConnectivityChanged = "net.connectivity.changed"
	TopicSMSReceived            = "sms.received"

	TopicOTAStatusChanged = "ota.status.changed"
	TopicOTAAvailable     = "ota.available"
	TopicOTAInstalled     = "ota.installed"

	TopicSettingsChanged = "settings.changed"
	TopicSystemBooted    = "system.booted"
	TopicSystemShutdown  = "system.shutdown"
	TopicDashboardReady  = "dashboard.ready"
)

// ButtonTopic builds the topic for a gesture on input-events, for example
// ButtonTopic("brake:left", "hold") is "button.brake.left.hold".
// The colon in the source name becomes a dot so the whole topic is one
// consistent namespace and globs like button.brake.* work.
func ButtonTopic(source, gesture string) string {
	return "button." + dotted(source) + "." + gesture
}

func dotted(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			out[i] = '.'
			continue
		}
		out[i] = s[i]
	}
	return string(out)
}
