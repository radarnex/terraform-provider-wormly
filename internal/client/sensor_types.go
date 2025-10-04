package client

// NOTE: These sensor type IDs are not officially documented in the Wormly API documentation.
// They have been determined from the API.
const (
	SensorTypePing    = "1"
	SensorTypeHTTP    = "2"
	SensorTypeUnknown = "3"
	SensorTypeSMTP    = "4"
	SensorTypePOP3    = "5"
	SensorTypeIMAP    = "6"
	SensorTypeFTP     = "7"
	SensorTypeTCP     = "8"
	SensorTypeDNS     = "9"
)

// SensorTypeNames provides a mapping from sensor type ID to human-readable name.
var SensorTypeNames = map[string]string{
	SensorTypePing:    "ping",
	SensorTypeHTTP:    "http",
	SensorTypeUnknown: "unknown",
	SensorTypeSMTP:    "smtp",
	SensorTypePOP3:    "pop3",
	SensorTypeIMAP:    "imap",
	SensorTypeFTP:     "ftp",
	SensorTypeTCP:     "tcp",
	SensorTypeDNS:     "dns",
}

// GetSensorTypeName returns the human-readable name for a sensor type ID.
// Returns "unknown" if the sensor type ID is not recognized.
func GetSensorTypeName(sensorTypeID string) string {
	if name, ok := SensorTypeNames[sensorTypeID]; ok {
		return name
	}
	return "unknown"
}
