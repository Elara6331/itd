package infinitime

import "tinygo.org/x/bluetooth"

type btChar struct {
	Name      string
	ID        bluetooth.UUID
	ServiceID bluetooth.UUID
}

var (
	musicServiceUUID      = mustParse("00000000-78fc-48fe-8e23-433b3a1942d0")
	navigationServiceUUID = mustParse("00010000-78fc-48fe-8e23-433b3a1942d0")
	motionServiceUUID     = mustParse("00030000-78fc-48fe-8e23-433b3a1942d0")
	weatherServiceUUID    = mustParse("00050000-78fc-48fe-8e23-433b3a1942d0")
)

var (
	newAlertChar = btChar{
		"New Alert",
		bluetooth.CharacteristicUUIDNewAlert,
		bluetooth.ServiceUUIDAlertNotification,
	}
	notifEventChar = btChar{
		"Notification Event",
		mustParse("00020001-78fc-48fe-8e23-433b3a1942d0"),
		bluetooth.ServiceUUIDAlertNotification,
	}
	stepCountChar = btChar{
		"Step Count",
		mustParse("00030001-78fc-48fe-8e23-433b3a1942d0"),
		motionServiceUUID,
	}
	rawMotionChar = btChar{
		"Raw Motion",
		mustParse("00030002-78fc-48fe-8e23-433b3a1942d0"),
		motionServiceUUID,
	}
	firmwareVerChar = btChar{
		"Firmware Version",
		bluetooth.CharacteristicUUIDFirmwareRevisionString,
		bluetooth.ServiceUUIDDeviceInformation,
	}
	currentTimeChar = btChar{
		"Current Time",
		bluetooth.CharacteristicUUIDCurrentTime,
		bluetooth.ServiceUUIDCurrentTime,
	}
	localTimeChar = btChar{
		"Local Time",
		bluetooth.CharacteristicUUIDLocalTimeInformation,
		bluetooth.ServiceUUIDCurrentTime,
	}
	batteryLevelChar = btChar{
		"Battery Level",
		bluetooth.CharacteristicUUIDBatteryLevel,
		bluetooth.ServiceUUIDBattery,
	}
	heartRateChar = btChar{
		"Heart Rate",
		bluetooth.CharacteristicUUIDHeartRateMeasurement,
		bluetooth.ServiceUUIDHeartRate,
	}
	fsVersionChar = btChar{
		"Filesystem Version",
		mustParse("adaf0200-4669-6c65-5472-616e73666572"),
		bluetooth.ServiceUUIDFileTransferByAdafruit,
	}
	fsTransferChar = btChar{
		"Filesystem Transfer",
		mustParse("adaf0200-4669-6c65-5472-616e73666572"),
		bluetooth.ServiceUUIDFileTransferByAdafruit,
	}
	dfuCtrlPointChar = btChar{
		"DFU Control Point",
		bluetooth.CharacteristicUUIDLegacyDFUControlPoint,
		bluetooth.ServiceUUIDLegacyDFU,
	}
	dfuPacketChar = btChar{
		"DFU Packet",
		bluetooth.CharacteristicUUIDLegacyDFUPacket,
		bluetooth.ServiceUUIDLegacyDFU,
	}
	navigationFlagsChar = btChar{
		"Navigation Flags",
		mustParse("00010001-78fc-48fe-8e23-433b3a1942d0"),
		navigationServiceUUID,
	}
	navigationNarrativeChar = btChar{
		"Navigation Narrative",
		mustParse("00010002-78fc-48fe-8e23-433b3a1942d0"),
		navigationServiceUUID,
	}
	navigationManDist = btChar{
		"Navigation Man Dist",
		mustParse("00010003-78fc-48fe-8e23-433b3a1942d0"),
		navigationServiceUUID,
	}
	navigationProgress = btChar{
		"Navigation Progress",
		mustParse("00010004-78fc-48fe-8e23-433b3a1942d0"),
		navigationServiceUUID,
	}
	weatherDataChar = btChar{
		"Weather Data",
		mustParse("00050001-78fc-48fe-8e23-433b3a1942d0"),
		weatherServiceUUID,
	}
	musicEventChar = btChar{
		"Music Event",
		mustParse("00000001-78fc-48fe-8e23-433b3a1942d0"),
		musicServiceUUID,
	}
	musicStatusChar = btChar{
		"Music Status",
		mustParse("00000002-78fc-48fe-8e23-433b3a1942d0"),
		musicServiceUUID,
	}
	musicArtistChar = btChar{
		"Music Artist",
		mustParse("00000003-78fc-48fe-8e23-433b3a1942d0"),
		musicServiceUUID,
	}
	musicTrackChar = btChar{
		"Music Track",
		mustParse("00000004-78fc-48fe-8e23-433b3a1942d0"),
		musicServiceUUID,
	}
	musicAlbumChar = btChar{
		"Music Album",
		mustParse("00000005-78fc-48fe-8e23-433b3a1942d0"),
		musicServiceUUID,
	}
)

func mustParse(s string) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}
