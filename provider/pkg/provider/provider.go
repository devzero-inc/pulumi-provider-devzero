package provider

import (
	"context"
	"errors"
	"os"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/resources"
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
// resolved. It applies env-var fallbacks, validates required fields, and
// initializes the package-level ClientSet.
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

	clientset.Set(NewClientSet(c.Token, c.TeamID, c.URL))
	return nil
}

// New constructs and returns the Pulumi DevZero provider.
// Resources are registered in infer.Options.Resources.
func New() p.Provider {
	return infer.Provider(infer.Options{
		Config: infer.Config(&ProviderConfig{}),
		Resources: []infer.InferredResource{
			infer.Resource[*resources.Cluster, resources.ClusterArgs, resources.ClusterState](&resources.Cluster{}),
		},
	})
}
