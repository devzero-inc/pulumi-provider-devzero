package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// ClusterArgs defines the user-configurable inputs for a Cluster resource.
type ClusterArgs struct {
	Name string `pulumi:"name"`
}

// ClusterState is the full persisted state: user inputs plus computed fields.
type ClusterState struct {
	ClusterArgs
	Token string `pulumi:"token,secret"`
}

// Annotate provides descriptions for SDK documentation and marks secret fields.
func (s *ClusterState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "The name of the cluster.")
	a.Describe(&s.Token, "Authentication token for the cluster. Rotated automatically if empty on update (e.g. after import).")
}

// Cluster is the resource implementation.
// It satisfies infer.CustomCreate, CustomRead, CustomUpdate, and CustomDelete.
type Cluster struct{}

// Create calls ClusterMutationService.CreateCluster and stores the cluster id and token.
func (c *Cluster) Create(ctx context.Context, req infer.CreateRequest[ClusterArgs]) (infer.CreateResponse[ClusterState], error) {
	if req.DryRun {
		return infer.CreateResponse[ClusterState]{Output: ClusterState{ClusterArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.CreateResponse[ClusterState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.ClusterMutationClient.CreateCluster(ctx, connect.NewRequest(&apiv1.CreateClusterRequest{
		TeamId:      cs.TeamID,
		ClusterName: req.Inputs.Name,
	}))
	if err != nil {
		return infer.CreateResponse[ClusterState]{}, fmt.Errorf("CreateCluster: %w", err)
	}
	if resp.Msg.Cluster == nil || resp.Msg.Token == "" {
		return infer.CreateResponse[ClusterState]{}, fmt.Errorf("CreateCluster: empty response from server")
	}

	return infer.CreateResponse[ClusterState]{
		ID: resp.Msg.Cluster.Id,
		Output: ClusterState{
			ClusterArgs: ClusterArgs{Name: req.Inputs.Name},
			Token:       resp.Msg.Token,
		},
	}, nil
}

// Read calls K8SService.GetCluster and refreshes the cluster name from the API.
// The token is NOT returned by GetCluster, so it is preserved from the existing state.
func (c *Cluster) Read(ctx context.Context, req infer.ReadRequest[ClusterArgs, ClusterState]) (infer.ReadResponse[ClusterArgs, ClusterState], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.ReadResponse[ClusterArgs, ClusterState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.K8SClient.GetCluster(ctx, connect.NewRequest(&apiv1.GetClusterRequest{
		TeamId:    cs.TeamID,
		ClusterId: req.ID,
	}))
	if err != nil {
		return infer.ReadResponse[ClusterArgs, ClusterState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetCluster: %w", err)
	}
	if resp.Msg.Cluster == nil {
		return infer.ReadResponse[ClusterArgs, ClusterState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetCluster: cluster not found")
	}

	// Prefer CustomName (user-set); fall back to the system-assigned Name.
	name := resp.Msg.Cluster.CustomName
	if name == "" {
		name = resp.Msg.Cluster.Name
	}

	updatedInputs := ClusterArgs{Name: name}
	return infer.ReadResponse[ClusterArgs, ClusterState]{
		ID:     req.ID,
		Inputs: updatedInputs,
		State: ClusterState{
			ClusterArgs: updatedInputs,
			Token:       req.State.Token, // token is not retrievable via Read — preserve it
		},
	}, nil
}

// Update calls ClusterMutationService.UpdateCluster.
// If the existing token is empty (e.g. after import), it also calls ResetClusterToken.
func (c *Cluster) Update(ctx context.Context, req infer.UpdateRequest[ClusterArgs, ClusterState]) (infer.UpdateResponse[ClusterState], error) {
	if req.DryRun {
		return infer.UpdateResponse[ClusterState]{Output: ClusterState{ClusterArgs: req.Inputs, Token: req.State.Token}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.UpdateResponse[ClusterState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	updateResp, err := cs.ClusterMutationClient.UpdateCluster(ctx, connect.NewRequest(&apiv1.UpdateClusterRequest{
		TeamId:      cs.TeamID,
		ClusterId:   req.ID,
		ClusterName: req.Inputs.Name,
	}))
	if err != nil {
		return infer.UpdateResponse[ClusterState]{}, fmt.Errorf("UpdateCluster: %w", err)
	}
	if updateResp.Msg.Cluster == nil {
		return infer.UpdateResponse[ClusterState]{}, fmt.Errorf("UpdateCluster: empty response from server")
	}

	token := req.State.Token
	// Rotate the token if it is empty — this happens after `pulumi import`.
	if token == "" {
		resetResp, err := cs.ClusterMutationClient.ResetClusterToken(ctx, connect.NewRequest(&apiv1.ResetClusterTokenRequest{
			TeamId:    cs.TeamID,
			ClusterId: req.ID,
		}))
		if err != nil {
			return infer.UpdateResponse[ClusterState]{}, fmt.Errorf("ResetClusterToken: %w", err)
		}
		if resetResp.Msg.Token == "" {
			return infer.UpdateResponse[ClusterState]{}, fmt.Errorf("ResetClusterToken: server returned empty token")
		}
		token = resetResp.Msg.Token
	}

	return infer.UpdateResponse[ClusterState]{
		Output: ClusterState{
			ClusterArgs: ClusterArgs{Name: updateResp.Msg.Cluster.CustomName},
			Token:       token,
		},
	}, nil
}

// Delete calls ClusterMutationService.DeleteCluster.
func (c *Cluster) Delete(ctx context.Context, req infer.DeleteRequest[ClusterState]) (infer.DeleteResponse, error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.DeleteResponse{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	_, err := cs.ClusterMutationClient.DeleteCluster(ctx, connect.NewRequest(&apiv1.DeleteClusterRequest{
		TeamId:    cs.TeamID,
		ClusterId: req.ID,
	}))
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("DeleteCluster: %w", err)
	}
	return infer.DeleteResponse{}, nil
}
