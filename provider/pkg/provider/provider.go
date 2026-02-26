package provider

import (
	"context"
	"errors"
	"os"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

const defaultURL = "https://dakr.devzero.io"

// ProviderConfig is the provider-level configuration.
// Token is marked secret so it is encrypted at rest in Pulumi state.
type ProviderConfig struct {
	Token  string `pulumi:"token,secret"`
	TeamID string `pulumi:"teamId"`
	URL    string `pulumi:"url,optional"`
}

// Annotate registers field descriptions and default env-var mappings used during
// schema generation and Pulumi config resolution.
func (c *ProviderConfig) Annotate(a infer.Annotator) {
	a.Describe(&c.Token, "DevZero API token. Can also be supplied via DEVZERO_TOKEN.")
	a.Describe(&c.TeamID, "DevZero team ID. Can also be supplied via DEVZERO_TEAM_ID.")
	a.Describe(&c.URL, "DevZero API base URL. Defaults to https://dakr.devzero.io.")
	a.SetDefault(&c.Token, "", "DEVZERO_TOKEN")
	a.SetDefault(&c.TeamID, "", "DEVZERO_TEAM_ID")
	a.SetDefault(&c.URL, defaultURL, "DEVZERO_URL")
}

// Configure is called by the Pulumi framework after all config values are
// resolved. It applies env-var fallbacks and validates required fields.
func (c *ProviderConfig) Configure(_ context.Context) error {
	if c.Token == "" {
		c.Token = os.Getenv("DEVZERO_TOKEN")
	}
	if c.TeamID == "" {
		c.TeamID = os.Getenv("DEVZERO_TEAM_ID")
	}
	if c.URL == "" {
		c.URL = os.Getenv("DEVZERO_URL")
	}
	if c.URL == "" {
		c.URL = defaultURL
	}

	if c.Token == "" {
		return errors.New("devzero: 'token' is required (set devzero:token or DEVZERO_TOKEN)")
	}
	if c.TeamID == "" {
		return errors.New("devzero: 'teamId' is required (set devzero:teamId or DEVZERO_TEAM_ID)")
	}

	activeClientSet = NewClientSet(c.Token, c.TeamID, c.URL)
	return nil
}

// Client is a DevZero API client stub (HTTP transport added later).
type Client struct {
	Token  string
	TeamID string
	URL    string
}

// NewClient builds a Client from validated provider configuration.
func NewClient(cfg ProviderConfig) *Client {
	return &Client{
		Token:  cfg.Token,
		TeamID: cfg.TeamID,
		URL:    cfg.URL,
	}
}

// New constructs and returns the Pulumi DevZero provider.
func New() p.Provider {
	return infer.Provider(infer.Options{
		Config: infer.Config(&ProviderConfig{}),
	})
}
