package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// GetClusterIDByNameArgs are the inputs for the GetClusterIdByName function.
type GetClusterIDByNameArgs struct {
	TeamID string `pulumi:"teamId"`
	Name   string `pulumi:"name"`
}

// GetClusterIDByNameResult is the output of the GetClusterIdByName function.
type GetClusterIDByNameResult struct {
	ClusterID string `pulumi:"clusterId"`
}

// Annotate provides descriptions used in SDK documentation.
func (a *GetClusterIDByNameArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.TeamID, "The team ID to search within.")
	ann.Describe(&a.Name, "The cluster name to look up.")
}

// Annotate provides descriptions used in SDK documentation.
func (r *GetClusterIDByNameResult) Annotate(ann infer.Annotator) {
	ann.Describe(&r.ClusterID, "The ID of the cluster matching the given team and name.")
}

// GetClusterIdByName is a Pulumi data source that looks up an existing cluster
// by team ID and name, returning its ID.
type GetClusterIdByName struct{}

// Invoke calls the backend ClusterService.GetClusterIDByName RPC.
func (f *GetClusterIdByName) Invoke(ctx context.Context, req infer.FunctionRequest[GetClusterIDByNameArgs]) (infer.FunctionResponse[GetClusterIDByNameResult], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.FunctionResponse[GetClusterIDByNameResult]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	teamID := req.Input.TeamID
	if teamID == "" {
		teamID = cs.TeamID
	}

	resp, err := cs.ClusterServiceClient.GetClusterIDByName(ctx, connect.NewRequest(&apiv1.GetClusterIDByNameRequest{
		TeamId: teamID,
		Name:   req.Input.Name,
	}))
	if err != nil {
		return infer.FunctionResponse[GetClusterIDByNameResult]{}, fmt.Errorf("GetClusterIDByName: %w", err)
	}
	if resp.Msg.Id == "" {
		return infer.FunctionResponse[GetClusterIDByNameResult]{}, fmt.Errorf("GetClusterIDByName: no cluster found with name %q in team %q", req.Input.Name, teamID)
	}

	return infer.FunctionResponse[GetClusterIDByNameResult]{
		Output: GetClusterIDByNameResult{ClusterID: resp.Msg.Id},
	}, nil
}