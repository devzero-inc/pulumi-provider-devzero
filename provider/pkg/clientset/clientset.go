package clientset

import (
	apiv1connect "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1/apiv1connect"
)

// ClientSet holds the three Connect service clients used by all resources.
type ClientSet struct {
	TeamID                string
	Token                 string
	ClusterMutationClient apiv1connect.ClusterMutationServiceClient
	K8SClient             apiv1connect.K8SServiceClient
	RecommendationClient  apiv1connect.K8SRecommendationServiceClient
}

// active is the package-level singleton set during ProviderConfig.Configure.
var active *ClientSet

// Set stores the initialized ClientSet. Called once by ProviderConfig.Configure.
func Set(cs *ClientSet) { active = cs }

// Get returns the initialized ClientSet. Returns nil before Configure is called.
func Get() *ClientSet { return active }
