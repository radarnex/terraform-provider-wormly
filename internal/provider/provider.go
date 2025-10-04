package provider

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
)

// Config represents the provider configuration.
type Config struct {
	APIKey            string
	BaseURL           string
	RequestsPerSecond float64
	MaxRetries        int
	InitialBackoff    time.Duration
	BackoffMultiplier float64
	MaxBackoff        time.Duration
	UserAgent         string
	Debug             bool
}

// wormlyProviderModel represents the provider configuration model.
type wormlyProviderModel struct {
	APIKey            types.String  `tfsdk:"api_key"`
	BaseURL           types.String  `tfsdk:"base_url"`
	RequestsPerSecond types.Float64 `tfsdk:"requests_per_second"`
	MaxRetries        types.Int64   `tfsdk:"max_retries"`
	InitialBackoff    types.String  `tfsdk:"initial_backoff"`
	BackoffMultiplier types.Float64 `tfsdk:"backoff_multiplier"`
	MaxBackoff        types.String  `tfsdk:"max_backoff"`
	UserAgent         types.String  `tfsdk:"user_agent"`
	Debug             types.Bool    `tfsdk:"debug"`
}

type wormlyProvider struct {
	version string
}

// New creates a new provider instance.
func New(version string) provider.Provider {
	return &wormlyProvider{
		version: version,
	}
}

func (p *wormlyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wormly"
	resp.Version = p.version
}

func (p *wormlyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Wormly API key.",
				Required:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for the Wormly API. Defaults to 'https://api.wormly.com'.",
				Optional:            true,
			},
			"requests_per_second": schema.Float64Attribute{
				MarkdownDescription: "Maximum number of requests per second to the Wormly API. Defaults to 10.",
				Optional:            true,
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retries for failed requests. Defaults to 3.",
				Optional:            true,
			},
			"initial_backoff": schema.StringAttribute{
				MarkdownDescription: "Initial backoff duration for retry attempts. Defaults to '1s'.",
				Optional:            true,
			},
			"backoff_multiplier": schema.Float64Attribute{
				MarkdownDescription: "Multiplier for exponential backoff. Defaults to 2.0.",
				Optional:            true,
			},
			"max_backoff": schema.StringAttribute{
				MarkdownDescription: "Maximum backoff duration. Defaults to '30s'.",
				Optional:            true,
			},
			"user_agent": schema.StringAttribute{
				MarkdownDescription: "User agent string for API requests. Defaults to 'terraform-provider-wormly/dev'.",
				Optional:            true,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Enable debug logging for API requests and responses. Defaults to false.",
				Optional:            true,
			},
		},
	}
}

func (p *wormlyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data wormlyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build configuration with defaults
	config := Config{
		APIKey:            data.APIKey.ValueString(),
		BaseURL:           "https://api.wormly.com",
		RequestsPerSecond: 3.0,
		MaxRetries:        3,
		InitialBackoff:    time.Second,
		BackoffMultiplier: 2.0,
		MaxBackoff:        30 * time.Second,
		UserAgent:         "terraform-provider-wormly/dev",
		Debug:             false,
	}

	// Override with configured values if provided
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		config.BaseURL = data.BaseURL.ValueString()
	}

	if !data.RequestsPerSecond.IsNull() && !data.RequestsPerSecond.IsUnknown() {
		config.RequestsPerSecond = data.RequestsPerSecond.ValueFloat64()
	}

	if !data.MaxRetries.IsNull() && !data.MaxRetries.IsUnknown() {
		config.MaxRetries = int(data.MaxRetries.ValueInt64())
	}

	if !data.InitialBackoff.IsNull() && !data.InitialBackoff.IsUnknown() {
		if duration, err := time.ParseDuration(data.InitialBackoff.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Initial Backoff Duration",
				"Could not parse initial_backoff as a duration: "+err.Error(),
			)
			return
		} else {
			config.InitialBackoff = duration
		}
	}

	if !data.BackoffMultiplier.IsNull() && !data.BackoffMultiplier.IsUnknown() {
		config.BackoffMultiplier = data.BackoffMultiplier.ValueFloat64()
	}

	if !data.MaxBackoff.IsNull() && !data.MaxBackoff.IsUnknown() {
		if duration, err := time.ParseDuration(data.MaxBackoff.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Max Backoff Duration",
				"Could not parse max_backoff as a duration: "+err.Error(),
			)
			return
		} else {
			config.MaxBackoff = duration
		}
	}

	if !data.UserAgent.IsNull() && !data.UserAgent.IsUnknown() {
		config.UserAgent = data.UserAgent.ValueString()
	}

	if !data.Debug.IsNull() && !data.Debug.IsUnknown() {
		config.Debug = data.Debug.ValueBool()
	}

	// Validate API key
	if config.APIKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"The api_key must be provided to authenticate with the Wormly API.",
		)
		return
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create logger for debug output
	var logger client.Logger = client.NoOpLogger{}
	if config.Debug {
		logger = client.NewStdLogger(log.New(os.Stderr, "[terraform-provider-wormly] ", log.LstdFlags))
	}

	// Create Wormly client
	wormlyClient, err := client.New(httpClient, config.APIKey, config.BaseURL, config.UserAgent,
		config.RequestsPerSecond, config.MaxRetries, config.InitialBackoff,
		config.BackoffMultiplier, config.MaxBackoff, logger, config.Debug)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Wormly API Client",
			"An unexpected error occurred when creating the Wormly API client: "+err.Error(),
		)
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = wormlyClient
	resp.ResourceData = wormlyClient
}

func (p *wormlyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
		NewSensorHTTPResource,
		NewGlobalAlertsMuteResource,
		NewScheduledDowntimePeriodResource,
	}
}

func (p *wormlyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHostDataSource,
		NewSensorHTTPDataSource,
	}
}
