package client

import (
	"testing"
)

func TestParseHTTPSensorParams(t *testing.T) {
	// Test JSON format
	jsonParams := `{
		"url": "https://example.com",
		"timeout": 30,
		"responsecode": "200",
		"verifysslcert": true,
		"expectedtext": "Welcome"
	}`

	params := parseHTTPSensorParams(jsonParams)

	if params.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %q", params.URL)
	}
	if params.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", params.Timeout)
	}
	if params.ResponseCode != "200" {
		t.Errorf("Expected response code '200', got %q", params.ResponseCode)
	}
	if !params.VerifySSLCert {
		t.Error("Expected VerifySSLCert to be true")
	}
	if params.ExpectedText != "Welcome" {
		t.Errorf("Expected ExpectedText 'Welcome', got %q", params.ExpectedText)
	}

	// Test key-value format
	kvParams := "url=https://test.com&timeout=60&responsecode=201&verifysslcert=1&expectedtext=Hello"

	params2 := parseHTTPSensorParams(kvParams)

	if params2.URL != "https://test.com" {
		t.Errorf("Expected URL 'https://test.com', got %q", params2.URL)
	}
	if params2.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", params2.Timeout)
	}
	if params2.ResponseCode != "201" {
		t.Errorf("Expected response code '201', got %q", params2.ResponseCode)
	}
	if !params2.VerifySSLCert {
		t.Error("Expected VerifySSLCert to be true")
	}
	if params2.ExpectedText != "Hello" {
		t.Errorf("Expected ExpectedText 'Hello', got %q", params2.ExpectedText)
	}
}

func TestConvertBasicSensorToHTTP(t *testing.T) {
	basicSensor := struct {
		HSID     string      `json:"hsid"`
		SensorID string      `json:"sensorid"`
		Enabled  string      `json:"enabled"`
		NiceName string      `json:"nicename"`
		Params   interface{} `json:"params"`
	}{
		HSID:     "123",
		SensorID: SensorTypeHTTP,
		Enabled:  "1",
		NiceName: "Test HTTP Sensor",
		Params: map[string]interface{}{
			"url":          "https://example.com",
			"timeout":      30,
			"responsecode": "200",
		},
	}

	httpSensor, err := convertBasicSensorToHTTP(basicSensor, 456)
	if err != nil {
		t.Fatalf("Failed to convert basic sensor: %v", err)
	}

	if httpSensor.ID != 123 {
		t.Errorf("Expected ID 123, got %d", httpSensor.ID)
	}
	if httpSensor.HostID != 456 {
		t.Errorf("Expected HostID 456, got %d", httpSensor.HostID)
	}
	if httpSensor.NiceName != "Test HTTP Sensor" {
		t.Errorf("Expected NiceName 'Test HTTP Sensor', got %q", httpSensor.NiceName)
	}
	if httpSensor.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %q", httpSensor.URL)
	}
	if httpSensor.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", httpSensor.Timeout)
	}
	if httpSensor.ResponseCode != "200" {
		t.Errorf("Expected response code '200', got %q", httpSensor.ResponseCode)
	}
}

func TestConvertBasicSensorToHTTP_InvalidHSID(t *testing.T) {
	basicSensor := struct {
		HSID     string      `json:"hsid"`
		SensorID string      `json:"sensorid"`
		Enabled  string      `json:"enabled"`
		NiceName string      `json:"nicename"`
		Params   interface{} `json:"params"`
	}{
		HSID:     "invalid_id",
		SensorID: SensorTypeHTTP,
		Enabled:  "1",
		NiceName: "Test HTTP Sensor",
		Params:   `{"url": "https://example.com"}`,
	}

	_, err := convertBasicSensorToHTTP(basicSensor, 456)
	if err == nil {
		t.Fatal("Expected error for invalid HSID, got nil")
	}

	expectedError := "invalid HSID value: invalid_id"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestConvertBasicSensorToHTTP_EnabledStringVariations(t *testing.T) {
	testCases := []struct {
		name         string
		enabledValue string
		description  string
	}{
		{"enabled_1", "1", "numeric true"},
		{"enabled_0", "0", "numeric false"},
		{"enabled_true", "true", "string true"},
		{"enabled_TRUE", "TRUE", "uppercase string true"},
		{"enabled_false", "false", "string false"},
		{"enabled_FALSE", "FALSE", "uppercase string false"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			basicSensor := struct {
				HSID     string      `json:"hsid"`
				SensorID string      `json:"sensorid"`
				Enabled  string      `json:"enabled"`
				NiceName string      `json:"nicename"`
				Params   interface{} `json:"params"`
			}{
				HSID:     "123",
				SensorID: SensorTypeHTTP,
				Enabled:  tc.enabledValue,
				NiceName: "Test HTTP Sensor",
				Params:   `{"url": "https://example.com"}`,
			}

			httpSensor, err := convertBasicSensorToHTTP(basicSensor, 456)
			if err != nil {
				t.Fatalf("Failed to convert basic sensor with %s: %v", tc.description, err)
			}

			if httpSensor.ID != 123 {
				t.Errorf("Expected ID 123, got %d", httpSensor.ID)
			}
			if httpSensor.HostID != 456 {
				t.Errorf("Expected HostID 456, got %d", httpSensor.HostID)
			}
		})
	}
}

func TestConvertBasicSensorToHTTP_ParamsTypes(t *testing.T) {
	// Test with JSON object params
	objectSensor := struct {
		HSID     string      `json:"hsid"`
		SensorID string      `json:"sensorid"`
		Enabled  string      `json:"enabled"`
		NiceName string      `json:"nicename"`
		Params   interface{} `json:"params"`
	}{
		HSID:     "123",
		SensorID: SensorTypeHTTP,
		Enabled:  "1",
		NiceName: "Test HTTP Sensor",
		Params: map[string]interface{}{
			"url":          "https://object-example.com",
			"timeout":      45,
			"responsecode": "201",
		},
	}

	httpSensor, err := convertBasicSensorToHTTP(objectSensor, 456)
	if err != nil {
		t.Fatalf("Failed to convert sensor with object params: %v", err)
	}

	if httpSensor.URL != "https://object-example.com" {
		t.Errorf("Expected URL 'https://object-example.com', got %q", httpSensor.URL)
	}
	if httpSensor.Timeout != 45 {
		t.Errorf("Expected timeout 45, got %d", httpSensor.Timeout)
	}

	// Test with JSON string params
	stringSensor := struct {
		HSID     string      `json:"hsid"`
		SensorID string      `json:"sensorid"`
		Enabled  string      `json:"enabled"`
		NiceName string      `json:"nicename"`
		Params   interface{} `json:"params"`
	}{
		HSID:     "124",
		SensorID: SensorTypeHTTP,
		Enabled:  "1",
		NiceName: "Test HTTP Sensor 2",
		Params:   `{"url": "https://string-example.com", "timeout": 60, "responsecode": "202"}`,
	}

	httpSensor2, err := convertBasicSensorToHTTP(stringSensor, 456)
	if err != nil {
		t.Fatalf("Failed to convert sensor with string params: %v", err)
	}

	if httpSensor2.URL != "https://string-example.com" {
		t.Errorf("Expected URL 'https://string-example.com', got %q", httpSensor2.URL)
	}
	if httpSensor2.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", httpSensor2.Timeout)
	}
}

func TestParseHTTPSensorParamsFromMap(t *testing.T) {
	paramsMap := map[string]interface{}{
		"url":                  "https://map-example.com",
		"timeout":              float64(120), // JSON numbers are often float64
		"responsecode":         "203",
		"verifysslcert":        true,
		"searchheaders":        false,
		"expectedtext":         "Success",
		"unwantedtext":         "Error",
		"sslvalidity":          float64(30),
		"cookies":              "session=abc123",
		"postparams":           "data=test",
		"customrequestheaders": "X-Custom: value",
		"useragent":            "TestAgent/1.0",
		"forceresolve":         "192.168.1.1",
	}

	params := parseHTTPSensorParamsFromMap(paramsMap)

	if params.URL != "https://map-example.com" {
		t.Errorf("Expected URL 'https://map-example.com', got %q", params.URL)
	}
	if params.Timeout != 120 {
		t.Errorf("Expected timeout 120, got %d", params.Timeout)
	}
	if params.ResponseCode != "203" {
		t.Errorf("Expected response code '203', got %q", params.ResponseCode)
	}
	if !params.VerifySSLCert {
		t.Error("Expected VerifySSLCert to be true")
	}
	if params.SearchHeaders {
		t.Error("Expected SearchHeaders to be false")
	}
	if params.ExpectedText != "Success" {
		t.Errorf("Expected ExpectedText 'Success', got %q", params.ExpectedText)
	}
	if params.UnwantedText != "Error" {
		t.Errorf("Expected UnwantedText 'Error', got %q", params.UnwantedText)
	}
	if params.SSLValidity != 30 {
		t.Errorf("Expected SSLValidity 30, got %d", params.SSLValidity)
	}
	if params.Cookies != "session=abc123" {
		t.Errorf("Expected Cookies 'session=abc123', got %q", params.Cookies)
	}
	if params.PostParams != "data=test" {
		t.Errorf("Expected PostParams 'data=test', got %q", params.PostParams)
	}
	if params.CustomRequestHeaders != "X-Custom: value" {
		t.Errorf("Expected CustomRequestHeaders 'X-Custom: value', got %q", params.CustomRequestHeaders)
	}
	if params.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected UserAgent 'TestAgent/1.0', got %q", params.UserAgent)
	}
	if params.ForceResolve != "192.168.1.1" {
		t.Errorf("Expected ForceResolve '192.168.1.1', got %q", params.ForceResolve)
	}
}

func TestConvertBasicSensorToHTTP_EnabledField(t *testing.T) {
	testCases := []struct {
		name          string
		enabledValue  string
		expectedValue bool
	}{
		{"enabled_1", "1", true},
		{"enabled_0", "0", false},
		{"enabled_true", "true", true},
		{"enabled_false", "false", false},
		{"enabled_TRUE", "TRUE", true},
		{"enabled_FALSE", "FALSE", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			basicSensor := struct {
				HSID     string      `json:"hsid"`
				SensorID string      `json:"sensorid"`
				Enabled  string      `json:"enabled"`
				NiceName string      `json:"nicename"`
				Params   interface{} `json:"params"`
			}{
				HSID:     "123",
				SensorID: SensorTypeHTTP,
				Enabled:  tc.enabledValue,
				NiceName: "Test HTTP Sensor",
				Params:   `{"url": "https://example.com"}`,
			}

			httpSensor, err := convertBasicSensorToHTTP(basicSensor, 456)
			if err != nil {
				t.Fatalf("Failed to convert basic sensor: %v", err)
			}

			if httpSensor.Enabled != tc.expectedValue {
				t.Errorf("Expected Enabled %v, got %v", tc.expectedValue, httpSensor.Enabled)
			}
		})
	}
}
