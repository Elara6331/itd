package infinitime

type NavFlag string

const (
	NavFlagArrive                  NavFlag = "arrive"
	NavFlagArriveLeft              NavFlag = "arrive-left"
	NavFlagArriveRight             NavFlag = "arrive-right"
	NavFlagArriveStraight          NavFlag = "arrive-straight"
	NavFlagClose                   NavFlag = "close"
	NavFlagContinue                NavFlag = "continue"
	NavFlagContinueLeft            NavFlag = "continue-left"
	NavFlagContinueRight           NavFlag = "continue-right"
	NavFlagContinueSlightLeft      NavFlag = "continue-slight-left"
	NavFlagContinueSlightRight     NavFlag = "continue-slight-right"
	NavFlagContinueStraight        NavFlag = "continue-straight"
	NavFlagContinueUturn           NavFlag = "continue-uturn"
	NavFlagDepart                  NavFlag = "depart"
	NavFlagDepartLeft              NavFlag = "depart-left"
	NavFlagDepartRight             NavFlag = "depart-right"
	NavFlagDepartStraight          NavFlag = "depart-straight"
	NavFlagEndOfRoadLeft           NavFlag = "end-of-road-left"
	NavFlagEndOfRoadRight          NavFlag = "end-of-road-right"
	NavFlagFerry                   NavFlag = "ferry"
	NavFlagFlag                    NavFlag = "flag"
	NavFlagFork                    NavFlag = "fork"
	NavFlagForkLeft                NavFlag = "fork-left"
	NavFlagForkRight               NavFlag = "fork-right"
	NavFlagForkSlightLeft          NavFlag = "fork-slight-left"
	NavFlagForkSlightRight         NavFlag = "fork-slight-right"
	NavFlagForkStraight            NavFlag = "fork-straight"
	NavFlagInvalid                 NavFlag = "invalid"
	NavFlagInvalidLeft             NavFlag = "invalid-left"
	NavFlagInvalidRight            NavFlag = "invalid-right"
	NavFlagInvalidSlightLeft       NavFlag = "invalid-slight-left"
	NavFlagInvalidSlightRight      NavFlag = "invalid-slight-right"
	NavFlagInvalidStraight         NavFlag = "invalid-straight"
	NavFlagInvalidUturn            NavFlag = "invalid-uturn"
	NavFlagMergeLeft               NavFlag = "merge-left"
	NavFlagMergeRight              NavFlag = "merge-right"
	NavFlagMergeSlightLeft         NavFlag = "merge-slight-left"
	NavFlagMergeSlightRight        NavFlag = "merge-slight-right"
	NavFlagMergeStraight           NavFlag = "merge-straight"
	NavFlagNewNameLeft             NavFlag = "new-name-left"
	NavFlagNewNameRight            NavFlag = "new-name-right"
	NavFlagNewNameSharpLeft        NavFlag = "new-name-sharp-left"
	NavFlagNewNameSharpRight       NavFlag = "new-name-sharp-right"
	NavFlagNewNameSlightLeft       NavFlag = "new-name-slight-left"
	NavFlagNewNameSlightRight      NavFlag = "new-name-slight-right"
	NavFlagNewNameStraight         NavFlag = "new-name-straight"
	NavFlagNotificationLeft        NavFlag = "notification-left"
	NavFlagNotificationRight       NavFlag = "notification-right"
	NavFlagNotificationSharpLeft   NavFlag = "notification-sharp-left"
	NavFlagNotificationSharpRight  NavFlag = "notification-sharp-right"
	NavFlagNotificationSlightLeft  NavFlag = "notification-slight-left"
	NavFlagNotificationSlightRight NavFlag = "notification-slight-right"
	NavFlagNotificationStraight    NavFlag = "notification-straight"
	NavFlagOffRampLeft             NavFlag = "off-ramp-left"
	NavFlagOffRampRight            NavFlag = "off-ramp-right"
	NavFlagOffRampSharpLeft        NavFlag = "off-ramp-sharp-left"
	NavFlagOffRampSharpRight       NavFlag = "off-ramp-sharp-right"
	NavFlagOffRampSlightLeft       NavFlag = "off-ramp-slight-left"
	NavFlagOffRampSlightRight      NavFlag = "off-ramp-slight-right"
	NavFlagOffRampStraight         NavFlag = "off-ramp-straight"
	NavFlagOnRampLeft              NavFlag = "on-ramp-left"
	NavFlagOnRampRight             NavFlag = "on-ramp-right"
	NavFlagOnRampSharpLeft         NavFlag = "on-ramp-sharp-left"
	NavFlagOnRampSharpRight        NavFlag = "on-ramp-sharp-right"
	NavFlagOnRampSlightLeft        NavFlag = "on-ramp-slight-left"
	NavFlagOnRampSlightRight       NavFlag = "on-ramp-slight-right"
	NavFlagOnRampStraight          NavFlag = "on-ramp-straight"
	NavFlagRotary                  NavFlag = "rotary"
	NavFlagRotaryLeft              NavFlag = "rotary-left"
	NavFlagRotaryRight             NavFlag = "rotary-right"
	NavFlagRotarySharpLeft         NavFlag = "rotary-sharp-left"
	NavFlagRotarySharpRight        NavFlag = "rotary-sharp-right"
	NavFlagRotarySlightLeft        NavFlag = "rotary-slight-left"
	NavFlagRotarySlightRight       NavFlag = "rotary-slight-right"
	NavFlagRotaryStraight          NavFlag = "rotary-straight"
	NavFlagRoundabout              NavFlag = "roundabout"
	NavFlagRoundaboutLeft          NavFlag = "roundabout-left"
	NavFlagRoundaboutRight         NavFlag = "roundabout-right"
	NavFlagRoundaboutSharpLeft     NavFlag = "roundabout-sharp-left"
	NavFlagRoundaboutSharpRight    NavFlag = "roundabout-sharp-right"
	NavFlagRoundaboutSlightLeft    NavFlag = "roundabout-slight-left"
	NavFlagRoundaboutSlightRight   NavFlag = "roundabout-slight-right"
	NavFlagRoundaboutStraight      NavFlag = "roundabout-straight"
	NavFlagTurnLeft                NavFlag = "turn-left"
	NavFlagTurnRight               NavFlag = "turn-right"
	NavFlagTurnSharpLeft           NavFlag = "turn-sharp-left"
	NavFlagTurnSharpRight          NavFlag = "turn-sharp-right"
	NavFlagTurnSlightLeft          NavFlag = "turn-slight-left"
	NavFlagTurnSlightRight         NavFlag = "turn-slight-right"
	NavFlagTurnStraight            NavFlag = "turn-straight"
	NavFlagUpDown                  NavFlag = "updown"
	NavFlagUTurn                   NavFlag = "uturn"
)

// SetNavFlag sets the navigation flag icon.
func (d *Device) SetNavFlag(flag NavFlag) error {
	char, err := d.getChar(navigationFlagsChar)
	if err != nil {
		return err
	}
	_, err = char.WriteWithoutResponse([]byte(flag))
	return err
}

// SetNavNarrative sets the navigation narrative string.
func (d *Device) SetNavNarrative(narrative string) error {
	char, err := d.getChar(navigationNarrativeChar)
	if err != nil {
		return err
	}
	_, err = char.WriteWithoutResponse([]byte(narrative))
	return err
}

// SetNavManeuverDistance sets the navigation maneuver distance.
func (d *Device) SetNavManeuverDistance(manDist string) error {
	char, err := d.getChar(navigationManDist)
	if err != nil {
		return err
	}
	_, err = char.WriteWithoutResponse([]byte(manDist))
	return err
}

// SetNavProgress sets the navigation progress.
func (d *Device) SetNavProgress(progress uint8) error {
	char, err := d.getChar(navigationProgress)
	if err != nil {
		return err
	}
	_, err = char.WriteWithoutResponse([]byte{progress})
	return err
}
