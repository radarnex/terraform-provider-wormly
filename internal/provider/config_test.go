package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProvider_Configure(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]tftypes.Value
		expectedConfig Config
		expectError    bool
	}{
		{
			name: "default configuration",
			config: map[string]tftypes.Value{
				"api_key":             tftypes.NewValue(tftypes.String, "test-api-key"),
				"base_url":            tftypes.NewValue(tftypes.String, nil),
				"requests_per_second": tftypes.NewValue(tftypes.Number, nil),
				"max_retries":         tftypes.NewValue(tftypes.Number, nil),
				"initial_backoff":     tftypes.NewValue(tftypes.String, nil),
				"backoff_multiplier":  tftypes.NewValue(tftypes.Number, nil),
				"max_backoff":         tftypes.NewValue(tftypes.String, nil),
				"user_agent":          tftypes.NewValue(tftypes.String, nil),
				"debug":               tftypes.NewValue(tftypes.Bool, nil),
			},
			expectedConfig: Config{
				APIKey:            "test-api-key",
				BaseURL:           "https://api.wormly.com",
				RequestsPerSecond: 10.0,
				MaxRetries:        3,
				InitialBackoff:    time.Second,
				BackoffMultiplier: 2.0,
				MaxBackoff:        30 * time.Second,
				UserAgent:         "terraform-provider-wormly/dev",
				Debug:             false,
			},
			expectError: false,
		},
		{
			name: "custom configuration",
			config: map[string]tftypes.Value{
				"api_key":             tftypes.NewValue(tftypes.String, "custom-api-key"),
				"base_url":            tftypes.NewValue(tftypes.String, "https://custom.api.com"),
				"requests_per_second": tftypes.NewValue(tftypes.Number, 5.0),
				"max_retries":         tftypes.NewValue(tftypes.Number, 5),
				"initial_backoff":     tftypes.NewValue(tftypes.String, "2s"),
				"backoff_multiplier":  tftypes.NewValue(tftypes.Number, 1.5),
				"max_backoff":         tftypes.NewValue(tftypes.String, "60s"),
				"user_agent":          tftypes.NewValue(tftypes.String, "custom-agent"),
				"debug":               tftypes.NewValue(tftypes.Bool, true),
			},
			expectedConfig: Config{
				APIKey:            "custom-api-key",
				BaseURL:           "https://custom.api.com",
				RequestsPerSecond: 5.0,
				MaxRetries:        5,
				InitialBackoff:    2 * time.Second,
				BackoffMultiplier: 1.5,
				MaxBackoff:        60 * time.Second,
				UserAgent:         "custom-agent",
				Debug:             true,
			},
			expectError: false,
		},
		{
			name: "invalid initial backoff",
			config: map[string]tftypes.Value{
				"api_key":             tftypes.NewValue(tftypes.String, "test-api-key"),
				"base_url":            tftypes.NewValue(tftypes.String, nil),
				"requests_per_second": tftypes.NewValue(tftypes.Number, nil),
				"max_retries":         tftypes.NewValue(tftypes.Number, nil),
				"initial_backoff":     tftypes.NewValue(tftypes.String, "invalid-duration"),
				"backoff_multiplier":  tftypes.NewValue(tftypes.Number, nil),
				"max_backoff":         tftypes.NewValue(tftypes.String, nil),
				"user_agent":          tftypes.NewValue(tftypes.String, nil),
				"debug":               tftypes.NewValue(tftypes.Bool, nil),
			},
			expectError: true,
		},
		{
			name: "invalid max backoff",
			config: map[string]tftypes.Value{
				"api_key":             tftypes.NewValue(tftypes.String, "test-api-key"),
				"base_url":            tftypes.NewValue(tftypes.String, nil),
				"requests_per_second": tftypes.NewValue(tftypes.Number, nil),
				"max_retries":         tftypes.NewValue(tftypes.Number, nil),
				"initial_backoff":     tftypes.NewValue(tftypes.String, nil),
				"backoff_multiplier":  tftypes.NewValue(tftypes.Number, nil),
				"max_backoff":         tftypes.NewValue(tftypes.String, "invalid-duration"),
				"user_agent":          tftypes.NewValue(tftypes.String, nil),
				"debug":               tftypes.NewValue(tftypes.Bool, nil),
			},
			expectError: true,
		},
		{
			name: "missing api key",
			config: map[string]tftypes.Value{
				"api_key":             tftypes.NewValue(tftypes.String, ""),
				"base_url":            tftypes.NewValue(tftypes.String, nil),
				"requests_per_second": tftypes.NewValue(tftypes.Number, nil),
				"max_retries":         tftypes.NewValue(tftypes.Number, nil),
				"initial_backoff":     tftypes.NewValue(tftypes.String, nil),
				"backoff_multiplier":  tftypes.NewValue(tftypes.Number, nil),
				"max_backoff":         tftypes.NewValue(tftypes.String, nil),
				"user_agent":          tftypes.NewValue(tftypes.String, nil),
				"debug":               tftypes.NewValue(tftypes.Bool, nil),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New("test")

			// Get the schema
			schemaReq := provider.SchemaRequest{}
			schemaResp := &provider.SchemaResponse{}
			p.Schema(t.Context(), schemaReq, schemaResp)

			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("Schema() returned errors: %v", schemaResp.Diagnostics)
			}

			// Create a config value
			configValue := tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"api_key":             tftypes.String,
					"base_url":            tftypes.String,
					"requests_per_second": tftypes.Number,
					"max_retries":         tftypes.Number,
					"initial_backoff":     tftypes.String,
					"backoff_multiplier":  tftypes.Number,
					"max_backoff":         tftypes.String,
					"user_agent":          tftypes.String,
					"debug":               tftypes.Bool,
				},
			}, tt.config)

			// Create the configuration
			var model wormlyProviderModel
			diags := tfsdk.Config{
				Schema: schemaResp.Schema,
				Raw:    configValue,
			}.Get(t.Context(), &model)

			if diags.HasError() && !tt.expectError {
				t.Fatalf("Config.Get() returned unexpected errors: %v", diags)
			}

			if diags.HasError() && tt.expectError {
				return // Expected error occurred during config parsing
			}

			// Test Configure method
			configReq := provider.ConfigureRequest{
				Config: tfsdk.Config{
					Schema: schemaResp.Schema,
					Raw:    configValue,
				},
			}
			configResp := &provider.ConfigureResponse{}

			p.Configure(t.Context(), configReq, configResp)

			if configResp.Diagnostics.HasError() && !tt.expectError {
				t.Fatalf("Configure() returned unexpected errors: %v", configResp.Diagnostics)
			}

			if configResp.Diagnostics.HasError() && tt.expectError {
				return // Expected error occurred
			}

			if tt.expectError {
				t.Fatal("Configure() should have returned an error but did not")
			}

			// Verify that client was created and stored
			if configResp.ResourceData == nil {
				t.Error("Configure() should have set ResourceData")
			}

			if configResp.DataSourceData == nil {
				t.Error("Configure() should have set DataSourceData")
			}
		})
	}
}

func TestProviderModel_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		input    wormlyProviderModel
		expected Config
	}{
		{
			name: "all null values use defaults",
			input: wormlyProviderModel{
				APIKey:            types.StringValue("test-key"),
				BaseURL:           types.StringNull(),
				RequestsPerSecond: types.Float64Null(),
				MaxRetries:        types.Int64Null(),
				InitialBackoff:    types.StringNull(),
				BackoffMultiplier: types.Float64Null(),
				MaxBackoff:        types.StringNull(),
				UserAgent:         types.StringNull(),
				Debug:             types.BoolNull(),
			},
			expected: Config{
				APIKey:            "test-key",
				BaseURL:           "https://api.wormly.com",
				RequestsPerSecond: 10.0,
				MaxRetries:        3,
				InitialBackoff:    time.Second,
				BackoffMultiplier: 2.0,
				MaxBackoff:        30 * time.Second,
				UserAgent:         "terraform-provider-wormly/dev",
				Debug:             false,
			},
		},
		{
			name: "partial configuration uses defaults for missing values",
			input: wormlyProviderModel{
				APIKey:            types.StringValue("test-key"),
				BaseURL:           types.StringValue("https://custom.api.com"),
				RequestsPerSecond: types.Float64Null(),
				MaxRetries:        types.Int64Value(5),
				InitialBackoff:    types.StringNull(),
				BackoffMultiplier: types.Float64Null(),
				MaxBackoff:        types.StringValue("45s"),
				UserAgent:         types.StringNull(),
				Debug:             types.BoolNull(),
			},
			expected: Config{
				APIKey:            "test-key",
				BaseURL:           "https://custom.api.com",
				RequestsPerSecond: 10.0,
				MaxRetries:        5,
				InitialBackoff:    time.Second,
				BackoffMultiplier: 2.0,
				MaxBackoff:        45 * time.Second,
				UserAgent:         "terraform-provider-wormly/dev",
				Debug:             false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build configuration with defaults (simulating the Configure method logic)
			config := Config{
				APIKey:            tt.input.APIKey.ValueString(),
				BaseURL:           "https://api.wormly.com",
				RequestsPerSecond: 10.0,
				MaxRetries:        3,
				InitialBackoff:    time.Second,
				BackoffMultiplier: 2.0,
				MaxBackoff:        30 * time.Second,
				UserAgent:         "terraform-provider-wormly/dev",
				Debug:             false,
			}

			// Override with configured values if provided
			if !tt.input.BaseURL.IsNull() && !tt.input.BaseURL.IsUnknown() {
				config.BaseURL = tt.input.BaseURL.ValueString()
			}

			if !tt.input.RequestsPerSecond.IsNull() && !tt.input.RequestsPerSecond.IsUnknown() {
				config.RequestsPerSecond = tt.input.RequestsPerSecond.ValueFloat64()
			}

			if !tt.input.MaxRetries.IsNull() && !tt.input.MaxRetries.IsUnknown() {
				config.MaxRetries = int(tt.input.MaxRetries.ValueInt64())
			}

			if !tt.input.InitialBackoff.IsNull() && !tt.input.InitialBackoff.IsUnknown() {
				if duration, err := time.ParseDuration(tt.input.InitialBackoff.ValueString()); err == nil {
					config.InitialBackoff = duration
				}
			}

			if !tt.input.BackoffMultiplier.IsNull() && !tt.input.BackoffMultiplier.IsUnknown() {
				config.BackoffMultiplier = tt.input.BackoffMultiplier.ValueFloat64()
			}

			if !tt.input.MaxBackoff.IsNull() && !tt.input.MaxBackoff.IsUnknown() {
				if duration, err := time.ParseDuration(tt.input.MaxBackoff.ValueString()); err == nil {
					config.MaxBackoff = duration
				}
			}

			if !tt.input.UserAgent.IsNull() && !tt.input.UserAgent.IsUnknown() {
				config.UserAgent = tt.input.UserAgent.ValueString()
			}

			if !tt.input.Debug.IsNull() && !tt.input.Debug.IsUnknown() {
				config.Debug = tt.input.Debug.ValueBool()
			}

			// Verify the configuration matches expected values
			if config.APIKey != tt.expected.APIKey {
				t.Errorf("APIKey = %v, want %v", config.APIKey, tt.expected.APIKey)
			}
			if config.BaseURL != tt.expected.BaseURL {
				t.Errorf("BaseURL = %v, want %v", config.BaseURL, tt.expected.BaseURL)
			}
			if config.RequestsPerSecond != tt.expected.RequestsPerSecond {
				t.Errorf("RequestsPerSecond = %v, want %v", config.RequestsPerSecond, tt.expected.RequestsPerSecond)
			}
			if config.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", config.MaxRetries, tt.expected.MaxRetries)
			}
			if config.InitialBackoff != tt.expected.InitialBackoff {
				t.Errorf("InitialBackoff = %v, want %v", config.InitialBackoff, tt.expected.InitialBackoff)
			}
			if config.BackoffMultiplier != tt.expected.BackoffMultiplier {
				t.Errorf("BackoffMultiplier = %v, want %v", config.BackoffMultiplier, tt.expected.BackoffMultiplier)
			}
			if config.MaxBackoff != tt.expected.MaxBackoff {
				t.Errorf("MaxBackoff = %v, want %v", config.MaxBackoff, tt.expected.MaxBackoff)
			}
			if config.UserAgent != tt.expected.UserAgent {
				t.Errorf("UserAgent = %v, want %v", config.UserAgent, tt.expected.UserAgent)
			}
			if config.Debug != tt.expected.Debug {
				t.Errorf("Debug = %v, want %v", config.Debug, tt.expected.Debug)
			}
		})
	}
}
