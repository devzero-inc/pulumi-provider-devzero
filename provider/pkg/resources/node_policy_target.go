package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// NodePolicyTargetArgs are the user-configurable inputs for a NodePolicyTarget resource.
type NodePolicyTargetArgs struct {
	Name        string   `pulumi:"name"`
	PolicyId    string   `pulumi:"policyId"`
	ClusterIds  []string `pulumi:"clusterIds"`
	Description *string  `pulumi:"description,optional"`
	Enabled     bool     `pulumi:"enabled,optional"`
}

// NodePolicyTargetState is the full persisted state (identical to args — no additional computed fields).
type NodePolicyTargetState struct {
	NodePolicyTargetArgs
}

// Annotate provides SDK documentation for NodePolicyTarget fields.
func (s *NodePolicyTargetState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "Human-friendly name for the target.")
	a.Describe(&s.PolicyId, "Node policy ID this target is attached to.")
	a.Describe(&s.ClusterIds, "Cluster IDs where this node policy applies.")
	a.Describe(&s.Description, "Free-form description of the target.")
	a.Describe(&s.Enabled, "Whether this target is active. Defaults to true.")
}

// NodePolicyTarget is the resource implementation.
type NodePolicyTarget struct{}

// ---------- CRUD ----------

func (n *NodePolicyTarget) Create(ctx context.Context, req infer.CreateRequest[NodePolicyTargetArgs]) (infer.CreateResponse[NodePolicyTargetState], error) {
	if req.DryRun {
		return infer.CreateResponse[NodePolicyTargetState]{Output: NodePolicyTargetState{NodePolicyTargetArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.CreateResponse[NodePolicyTargetState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.CreateNodePolicyTargets(ctx, connect.NewRequest(&apiv1.CreateNodePolicyTargetsRequest{
		Targets: []*apiv1.NodePolicyTarget{nodePolicyTargetArgsToProto(cs.TeamID, "", req.Inputs)},
	}))
	if err != nil {
		return infer.CreateResponse[NodePolicyTargetState]{}, fmt.Errorf("CreateNodePolicyTargets: %w", err)
	}
	if len(resp.Msg.Targets) == 0 {
		return infer.CreateResponse[NodePolicyTargetState]{}, fmt.Errorf("CreateNodePolicyTargets: empty response from server")
	}

	created := resp.Msg.Targets[0]
	return infer.CreateResponse[NodePolicyTargetState]{
		ID:     created.TargetId,
		Output: NodePolicyTargetState{NodePolicyTargetArgs: nodePolicyTargetProtoToArgs(created)},
	}, nil
}

func (n *NodePolicyTarget) Read(ctx context.Context, req infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]) (infer.ReadResponse[NodePolicyTargetArgs, NodePolicyTargetState], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.ReadResponse[NodePolicyTargetArgs, NodePolicyTargetState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.ListNodePolicyTargets(ctx, connect.NewRequest(&apiv1.ListNodePolicyTargetsRequest{
		TeamId: cs.TeamID,
	}))
	if err != nil {
		return infer.ReadResponse[NodePolicyTargetArgs, NodePolicyTargetState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("ListNodePolicyTargets: %w", err)
	}

	for _, t := range resp.Msg.Targets {
		if t.TargetId == req.ID {
			updatedArgs := nodePolicyTargetProtoToArgs(t)
			return infer.ReadResponse[NodePolicyTargetArgs, NodePolicyTargetState]{
				ID:     req.ID,
				Inputs: updatedArgs,
				State:  NodePolicyTargetState{NodePolicyTargetArgs: updatedArgs},
			}, nil
		}
	}

	return infer.ReadResponse[NodePolicyTargetArgs, NodePolicyTargetState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
		fmt.Errorf("ListNodePolicyTargets: target %q not found", req.ID)
}

func (n *NodePolicyTarget) Update(ctx context.Context, req infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]) (infer.UpdateResponse[NodePolicyTargetState], error) {
	if req.DryRun {
		return infer.UpdateResponse[NodePolicyTargetState]{Output: NodePolicyTargetState{NodePolicyTargetArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.UpdateResponse[NodePolicyTargetState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.UpdateNodePolicyTarget(ctx, connect.NewRequest(&apiv1.UpdateNodePolicyTargetRequest{
		Target: nodePolicyTargetArgsToProto(cs.TeamID, req.ID, req.Inputs),
	}))
	if err != nil {
		return infer.UpdateResponse[NodePolicyTargetState]{}, fmt.Errorf("UpdateNodePolicyTarget: %w", err)
	}
	if resp.Msg.Target == nil {
		return infer.UpdateResponse[NodePolicyTargetState]{}, fmt.Errorf("UpdateNodePolicyTarget: empty response from server")
	}

	return infer.UpdateResponse[NodePolicyTargetState]{
		Output: NodePolicyTargetState{NodePolicyTargetArgs: nodePolicyTargetProtoToArgs(resp.Msg.Target)},
	}, nil
}

// Delete removes the resource from Pulumi state only — no delete endpoint exists for NodePolicyTarget.
func (n *NodePolicyTarget) Delete(_ context.Context, _ infer.DeleteRequest[NodePolicyTargetState]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}

// ---------- proto conversion ----------

func nodePolicyTargetArgsToProto(teamID, id string, a NodePolicyTargetArgs) *apiv1.NodePolicyTarget {
	t := &apiv1.NodePolicyTarget{
		TargetId:   id,
		TeamId:     teamID,
		Name:       a.Name,
		PolicyId:   a.PolicyId,
		ClusterIds: a.ClusterIds,
		Enabled:    a.Enabled,
	}
	if a.Description != nil {
		t.Description = *a.Description
	}
	return t
}

func nodePolicyTargetProtoToArgs(t *apiv1.NodePolicyTarget) NodePolicyTargetArgs {
	a := NodePolicyTargetArgs{
		Name:       t.Name,
		PolicyId:   t.PolicyId,
		ClusterIds: t.ClusterIds,
		Enabled:    t.Enabled,
	}
	if t.Description != "" {
		a.Description = &t.Description
	}
	return a
}
