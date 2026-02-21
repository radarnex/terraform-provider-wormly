package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SensorHTTP represents a Wormly HTTP sensor.
type SensorHTTP struct {
	ID                   int       `json:"id"`
	HostID               int       `json:"hostid"`
	URL                  string    `json:"url"`
	NiceName             string    `json:"nicename"`
	Enabled              bool      `json:"enabled"`
	Timeout              int       `json:"timeout"`
	ResponseCode         string    `json:"responsecode"`
	VerifySSLCert        bool      `json:"verifysslcert"`
	SearchHeaders        bool      `json:"searchheaders"`
	ExpectedText         string    `json:"expectedtext"`
	UnwantedText         string    `json:"unwantedtext"`
	SSLValidity          int       `json:"sslvalidity"`
	Cookies              string    `json:"cookies"`
	PostParams           string    `json:"postparams"`
	CustomRequestHeaders string    `json:"customrequestheaders"`
	UserAgent            string    `json:"useragent"`
	ForceResolve         string    `json:"forceresolve"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// SensorHTTPCreateRequest represents the request payload for creating an HTTP sensor.
type SensorHTTPCreateRequest struct {
	HostID               int    `json:"hostid"`
	URL                  string `json:"url"`
	NiceName             string `json:"nicename,omitempty"`
	Timeout              int    `json:"timeout,omitempty"`
	ResponseCode         string `json:"responsecode,omitempty"`
	VerifySSLCert        bool   `json:"verifysslcert,omitempty"`
	SearchHeaders        bool   `json:"searchheaders,omitempty"`
	ExpectedText         string `json:"expectedtext,omitempty"`
	UnwantedText         string `json:"unwantedtext,omitempty"`
	SSLValidity          int    `json:"sslvalidity,omitempty"`
	Cookies              string `json:"cookies,omitempty"`
	PostParams           string `json:"postparams,omitempty"`
	CustomRequestHeaders string `json:"customrequestheaders,omitempty"`
	UserAgent            string `json:"useragent,omitempty"`
	ForceResolve         string `json:"forceresolve,omitempty"`
}

// WormlyHTTPSensorResponse represents the API response for HTTP sensor operations.
type WormlyHTTPSensorResponse struct {
	ErrorCode    int    `json:"errorcode"`
	Message      string `json:"message,omitempty"`
	HostSensorID int    `json:"hostsensorid,omitempty"`
}

// WormlyHTTPSensorListResponse represents the API response for getHostSensors.
type WormlyHTTPSensorListResponse struct {
	ErrorCode int `json:"errorcode"`
	Sensors   []struct {
		HSID     string      `json:"hsid"`     // The HostSensorID of the sensor (returned as string)
		SensorID string      `json:"sensorid"` // The ID of the sensor type (returned as string)
		Enabled  string      `json:"enabled"`  // Whether this sensor is enabled for testing (returned as string)
		NiceName string      `json:"nicename"` // The (optional) nicename for this sensor (API docs incorrectly say "nickname", actual response uses "nicename")
		Params   interface{} `json:"params"`   // Sensor parameters (can be object or string)
	} `json:"sensors"`
}

// SensorHTTPAPI defines the interface for HTTP sensor-related operations.
type SensorHTTPAPI interface {
	CreateSensorHTTP(ctx context.Context, req *SensorHTTPCreateRequest) (*SensorHTTP, error)
	GetSensorHTTP(ctx context.Context, hostID, sensorID int) (*SensorHTTP, error)
	DeleteSensorHTTP(ctx context.Context, sensorID int) error
	ListSensorHTTP(ctx context.Context, hostID int) ([]*SensorHTTP, error)
	EnableSensorHTTP(ctx context.Context, hsid int) error
	DisableSensorHTTP(ctx context.Context, hsid int) error
}

// Ensure Client implements SensorHTTPAPI.
var _ SensorHTTPAPI = (*Client)(nil)

// CreateSensorHTTP creates a new HTTP sensor.
func (c *Client) CreateSensorHTTP(ctx context.Context, req *SensorHTTPCreateRequest) (*SensorHTTP, error) {
	params := map[string]string{
		"hostid": strconv.Itoa(req.HostID),
		"url":    req.URL,
	}

	// Add optional parameters
	if req.NiceName != "" {
		params["nicename"] = req.NiceName
	}
	if req.Timeout > 0 {
		params["timeout"] = strconv.Itoa(req.Timeout)
	}
	if req.ResponseCode != "" {
		params["responsecode"] = req.ResponseCode
	}
	if req.VerifySSLCert {
		params["verifysslcert"] = "1"
	} else {
		params["verifysslcert"] = "0"
	}
	if req.SearchHeaders {
		params["searchheaders"] = "1"
	} else {
		params["searchheaders"] = "0"
	}
	if req.ExpectedText != "" {
		params["expectedtext"] = req.ExpectedText
	}
	if req.UnwantedText != "" {
		params["unwantedtext"] = req.UnwantedText
	}
	if req.SSLValidity > 0 {
		params["sslvalidity"] = strconv.Itoa(req.SSLValidity)
	}
	if req.Cookies != "" {
		params["cookies"] = req.Cookies
	}
	if req.PostParams != "" {
		params["postparams"] = req.PostParams
	}
	if req.CustomRequestHeaders != "" {
		params["customrequestheaders"] = req.CustomRequestHeaders
	}
	if req.UserAgent != "" {
		params["useragent"] = req.UserAgent
	}
	if req.ForceResolve != "" {
		params["forceresolve"] = req.ForceResolve
	}

	var response WormlyHTTPSensorResponse
	if err := c.makeFormRequest(ctx, "addHostSensor_HTTP", params, &response); err != nil {
		return nil, fmt.Errorf("failed to create HTTP sensor: %w", err)
	}

	if response.ErrorCode != 0 {
		return nil, fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return &SensorHTTP{
		ID:                   response.HostSensorID,
		HostID:               req.HostID,
		URL:                  req.URL,
		NiceName:             req.NiceName,
		Enabled:              true, // Sensors are created enabled by default according to Wormly API
		Timeout:              req.Timeout,
		ResponseCode:         req.ResponseCode,
		VerifySSLCert:        req.VerifySSLCert,
		SearchHeaders:        req.SearchHeaders,
		ExpectedText:         req.ExpectedText,
		UnwantedText:         req.UnwantedText,
		SSLValidity:          req.SSLValidity,
		Cookies:              req.Cookies,
		PostParams:           req.PostParams,
		CustomRequestHeaders: req.CustomRequestHeaders,
		UserAgent:            req.UserAgent,
		ForceResolve:         req.ForceResolve,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}

// GetSensorHTTP retrieves an HTTP sensor by host ID and sensor ID.
func (c *Client) GetSensorHTTP(ctx context.Context, hostID, sensorID int) (*SensorHTTP, error) {
	params := map[string]string{
		"hostid": strconv.Itoa(hostID),
	}

	var response WormlyHTTPSensorListResponse
	if err := c.makeFormRequest(ctx, "getHostSensors", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get HTTP sensor: %w", err)
	}

	if response.ErrorCode != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.ErrorCode)
	}

	// Find the specific sensor by HSID (HostSensorID)
	for _, sensor := range response.Sensors {
		// Convert string HSID to int for comparison
		hsid, err := strconv.Atoi(sensor.HSID)
		if err != nil {
			continue // Skip sensors with invalid HSID
		}
		if hsid == sensorID {
			return convertBasicSensorToHTTP(sensor, hostID)
		}
	}

	return nil, fmt.Errorf("HTTP sensor with ID %d not found for host %d", sensorID, hostID)
}

// DeleteSensorHTTP deletes an HTTP sensor by ID.
// Note: The sensorID parameter should be the HSID (HostSensorID) value.
func (c *Client) DeleteSensorHTTP(ctx context.Context, sensorID int) error {
	params := map[string]string{
		"hsid": strconv.Itoa(sensorID), // API expects hsid (HostSensorID)
	}

	var response WormlyHTTPSensorResponse
	if err := c.makeFormRequest(ctx, "deleteSensor", params, &response); err != nil {
		return fmt.Errorf("failed to delete HTTP sensor: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}

// ListSensorHTTP lists all HTTP sensors for a given host ID.
func (c *Client) ListSensorHTTP(ctx context.Context, hostID int) ([]*SensorHTTP, error) {
	params := map[string]string{
		"hostid": strconv.Itoa(hostID),
	}

	var response WormlyHTTPSensorListResponse
	if err := c.makeFormRequest(ctx, "getHostSensors", params, &response); err != nil {
		return nil, fmt.Errorf("failed to list HTTP sensors: %w", err)
	}

	if response.ErrorCode != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.ErrorCode)
	}

	var httpSensors []*SensorHTTP
	for _, sensor := range response.Sensors {
		if sensor.SensorID != SensorTypeHTTP {
			continue
		}

		httpSensor, err := convertBasicSensorToHTTP(sensor, hostID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert sensor (HSID: %s): %w", sensor.HSID, err)
		}
		httpSensors = append(httpSensors, httpSensor)
	}

	return httpSensors, nil
}

// EnableSensorHTTP enables an HTTP sensor by HSID.
func (c *Client) EnableSensorHTTP(ctx context.Context, hsid int) error {
	params := map[string]string{
		"hsid": strconv.Itoa(hsid),
	}

	var response WormlyHTTPSensorResponse
	if err := c.makeFormRequest(ctx, "enableSensor", params, &response); err != nil {
		return fmt.Errorf("failed to enable HTTP sensor: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}

// DisableSensorHTTP disables an HTTP sensor by HSID.
func (c *Client) DisableSensorHTTP(ctx context.Context, hsid int) error {
	params := map[string]string{
		"hsid": strconv.Itoa(hsid),
	}

	var response WormlyHTTPSensorResponse
	if err := c.makeFormRequest(ctx, "disableSensor", params, &response); err != nil {
		return fmt.Errorf("failed to disable HTTP sensor: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}

// HTTPSensorParams represents the parsed parameters from the sensor params field.
type HTTPSensorParams struct {
	URL                  string `json:"url"`
	Timeout              int    `json:"timeout"`
	ResponseCode         string `json:"responsecode"`
	VerifySSLCert        bool   `json:"verifysslcert"`
	SearchHeaders        bool   `json:"searchheaders"`
	ExpectedText         string `json:"expectedtext"`
	UnwantedText         string `json:"unwantedtext"`
	SSLValidity          int    `json:"sslvalidity"`
	Cookies              string `json:"cookies"`
	PostParams           string `json:"postparams"`
	CustomRequestHeaders string `json:"customrequestheaders"`
	UserAgent            string `json:"useragent"`
	ForceResolve         string `json:"forceresolve"`
}

// parseHTTPSensorParams parses the params string to extract HTTP sensor configuration.
func parseHTTPSensorParams(paramsStr string) *HTTPSensorParams {
	// The params field might be JSON or key-value pairs
	// Try JSON first
	var params HTTPSensorParams
	if err := json.Unmarshal([]byte(paramsStr), &params); err == nil {
		return &params
	}

	// If JSON parsing fails, try parsing as key-value pairs
	// This assumes params are in format "key1=value1&key2=value2" or similar
	params = HTTPSensorParams{}
	pairs := strings.Split(paramsStr, "&")
	for _, pair := range pairs {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			switch key {
			case "url":
				params.URL = value
			case "timeout":
				if timeout, err := strconv.Atoi(value); err == nil {
					params.Timeout = timeout
				}
			case "responsecode":
				params.ResponseCode = value
			case "verifysslcert":
				params.VerifySSLCert = value == "1" || strings.ToLower(value) == "true"
			case "searchheaders":
				params.SearchHeaders = value == "1" || strings.ToLower(value) == "true"
			case "expectedtext":
				params.ExpectedText = value
			case "unwantedtext":
				params.UnwantedText = value
			case "sslvalidity":
				if validity, err := strconv.Atoi(value); err == nil {
					params.SSLValidity = validity
				}
			case "cookies":
				params.Cookies = value
			case "postparams":
				params.PostParams = value
			case "customrequestheaders":
				params.CustomRequestHeaders = value
			case "useragent":
				params.UserAgent = value
			case "forceresolve":
				params.ForceResolve = value
			}
		}
	}

	return &params
}

// parseHTTPSensorParamsFromMap parses HTTP sensor parameters from a map.
func parseHTTPSensorParamsFromMap(paramsMap map[string]interface{}) *HTTPSensorParams {
	params := &HTTPSensorParams{}

	if url, ok := paramsMap["url"].(string); ok {
		params.URL = url
	}

	if timeout, ok := paramsMap["timeout"].(string); ok {
		if t, err := strconv.Atoi(timeout); err == nil {
			params.Timeout = t
		}
	} else if timeout, ok := paramsMap["timeout"].(float64); ok {
		params.Timeout = int(timeout)
	} else if timeout, ok := paramsMap["timeout"].(int); ok {
		params.Timeout = timeout
	}

	if responseCode, ok := paramsMap["responsecode"].(string); ok {
		params.ResponseCode = responseCode
	}

	// API uses "ssl_strict" instead of "verifysslcert"
	if sslStrict, ok := paramsMap["ssl_strict"].(string); ok {
		params.VerifySSLCert = sslStrict == "1" || strings.ToLower(sslStrict) == "true"
	} else if verifySsl, ok := paramsMap["verifysslcert"].(bool); ok {
		params.VerifySSLCert = verifySsl
	} else if verifySsl, ok := paramsMap["verifysslcert"].(string); ok {
		params.VerifySSLCert = verifySsl == "1" || strings.ToLower(verifySsl) == "true"
	}

	if searchHeaders, ok := paramsMap["searchheaders"].(string); ok {
		params.SearchHeaders = searchHeaders == "1" || strings.ToLower(searchHeaders) == "true"
	} else if searchHeaders, ok := paramsMap["searchheaders"].(bool); ok {
		params.SearchHeaders = searchHeaders
	}

	// API uses "wantedstring" instead of "expectedtext"
	if wantedString, ok := paramsMap["wantedstring"].(string); ok {
		params.ExpectedText = wantedString
	} else if expectedText, ok := paramsMap["expectedtext"].(string); ok {
		params.ExpectedText = expectedText
	}

	if unwantedText, ok := paramsMap["unwantedtext"].(string); ok {
		params.UnwantedText = unwantedText
	}

	if sslMinExpiry, ok := paramsMap["ssl_min_expiry_in"].(string); ok {
		if s, err := strconv.Atoi(sslMinExpiry); err == nil {
			params.SSLValidity = s
		}
	} else if sslMinExpiry, ok := paramsMap["ssl_min_expiry_in"].(float64); ok {
		params.SSLValidity = int(sslMinExpiry)
	} else if sslMinExpiry, ok := paramsMap["ssl_min_expiry_in"].(int); ok {
		params.SSLValidity = sslMinExpiry
	} else if sslValidity, ok := paramsMap["sslvalidity"].(string); ok {
		if s, err := strconv.Atoi(sslValidity); err == nil {
			params.SSLValidity = s
		}
	} else if sslValidity, ok := paramsMap["sslvalidity"].(float64); ok {
		params.SSLValidity = int(sslValidity)
	} else if sslValidity, ok := paramsMap["sslvalidity"].(int); ok {
		params.SSLValidity = sslValidity
	}

	if cookies, ok := paramsMap["cookies"].(string); ok {
		params.Cookies = cookies
	}

	if postParams, ok := paramsMap["postparams"].(string); ok {
		params.PostParams = postParams
	}

	if customHeaders, ok := paramsMap["customrequestheaders"].(string); ok {
		params.CustomRequestHeaders = customHeaders
	}

	if userAgent, ok := paramsMap["useragent"].(string); ok {
		params.UserAgent = userAgent
	}

	if forceResolve, ok := paramsMap["forceresolve"].(string); ok {
		params.ForceResolve = forceResolve
	}

	return params
}

// convertBasicSensorToHTTP converts a basic sensor from getHostSensors to a full SensorHTTP struct.
func convertBasicSensorToHTTP(sensor struct {
	HSID     string      `json:"hsid"`
	SensorID string      `json:"sensorid"`
	Enabled  string      `json:"enabled"`
	NiceName string      `json:"nicename"` // API docs incorrectly say "nickname", actual response uses "nicename"
	Params   interface{} `json:"params"`
}, hostID int) (*SensorHTTP, error) {
	// Convert HSID from string to int
	hsid, hsidErr := strconv.Atoi(sensor.HSID)
	if hsidErr != nil {
		return nil, fmt.Errorf("invalid HSID value: %s", sensor.HSID)
	}

	// Parse the enabled field - API returns string values like "1", "0", "true", "false"
	enabled := false
	switch strings.ToLower(sensor.Enabled) {
	case "1", "true":
		enabled = true
	case "0", "false":
		enabled = false
	}

	// Convert Params to string for parsing
	var httpParams *HTTPSensorParams

	switch p := sensor.Params.(type) {
	case string:
		httpParams = parseHTTPSensorParams(p)
	case map[string]interface{}:
		// Parse directly from map for better type handling
		httpParams = parseHTTPSensorParamsFromMap(p)
	case nil:
		httpParams = &HTTPSensorParams{}
	default:
		// Try to marshal whatever type it is and parse as JSON
		jsonBytes, marshalErr := json.Marshal(p)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal params of type %T: %w", p, marshalErr)
		}
		httpParams = parseHTTPSensorParams(string(jsonBytes))
	}

	return &SensorHTTP{
		ID:                   hsid,
		HostID:               hostID,
		URL:                  httpParams.URL,
		NiceName:             sensor.NiceName, // Fixed field reference
		Enabled:              enabled,
		Timeout:              httpParams.Timeout,
		ResponseCode:         httpParams.ResponseCode,
		VerifySSLCert:        httpParams.VerifySSLCert,
		SearchHeaders:        httpParams.SearchHeaders,
		ExpectedText:         httpParams.ExpectedText,
		UnwantedText:         httpParams.UnwantedText,
		SSLValidity:          httpParams.SSLValidity,
		Cookies:              httpParams.Cookies,
		PostParams:           httpParams.PostParams,
		CustomRequestHeaders: httpParams.CustomRequestHeaders,
		UserAgent:            httpParams.UserAgent,
		ForceResolve:         httpParams.ForceResolve,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}
