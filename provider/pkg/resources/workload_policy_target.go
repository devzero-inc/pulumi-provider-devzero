package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	commonv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// NamePatternArgs matches workloads by name using a regular expression.
type NamePatternArgs struct {
	Pattern string  `pulumi:"pattern,optional"`
	Flags   *string `pulumi:"flags,optional"`
}

// Annotate provides SDK documentation for NamePatternArgs fields.
func (n *NamePatternArgs) Annotate(a infer.Annotator) {
	a.Describe(&n.Pattern, "Regular expression pattern to match workload names.")
	a.Describe(&n.Flags, "Optional regex flags (e.g. 'i' for case-insensitive).")
}

// LabelSelectorRequirementArgs is a single match expression in a label selector.
type LabelSelectorRequirementArgs struct {
	Key      string   `pulumi:"key,optional"`
	Operator string   `pulumi:"operator,optional"`
	Values   []string `pulumi:"values,optional"`
}

// Annotate provides SDK documentation for LabelSelectorRequirementArgs fields.
func (l *LabelSelectorRequirementArgs) Annotate(a infer.Annotator) {
	a.Describe(&l.Key, "The label key that the selector applies to.")
	a.Describe(&l.Operator, "Operator relating the key and values. One of: 'In', 'NotIn', 'Exists', 'DoesNotExist'.")
	a.Describe(&l.Values, "Array of string values. Required for 'In' and 'NotIn' operators.")
}

// LabelSelectorArgs selects objects by labels (matchLabels and/or matchExpressions).
type LabelSelectorArgs struct {
	MatchLabels      map[string]string              `pulumi:"matchLabels,optional"`
	MatchExpressions []LabelSelectorRequirementArgs `pulumi:"matchExpressions,optional"`
}

// Annotate provides SDK documentation for LabelSelectorArgs fields.
func (l *LabelSelectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&l.MatchLabels, "Map of label key-value pairs that must all match exactly.")
	a.Describe(&l.MatchExpressions, "List of label selector requirements joined by AND.")
}

// WorkloadPolicyTargetArgs are the user-configurable inputs for a WorkloadPolicyTarget resource.
type WorkloadPolicyTargetArgs struct {
	Name              string             `pulumi:"name"`
	PolicyId          string             `pulumi:"policyId"`
	ClusterIds        []string           `pulumi:"clusterIds"`
	Description       *string            `pulumi:"description,optional"`
	Priority          int                `pulumi:"priority,optional"`
	Enabled           bool               `pulumi:"enabled,optional"`
	WorkloadNames     []string           `pulumi:"workloadNames,optional"`
	NodeGroupNames    []string           `pulumi:"nodeGroupNames,optional"`
	KindFilter        []string           `pulumi:"kindFilter,optional"`
	NamePattern       *NamePatternArgs   `pulumi:"namePattern,optional"`
	NamespaceSelector *LabelSelectorArgs `pulumi:"namespaceSelector,optional"`
	WorkloadSelector  *LabelSelectorArgs `pulumi:"workloadSelector,optional"`
}

// WorkloadPolicyTargetState is the full persisted state (identical to args — no additional computed fields).
type WorkloadPolicyTargetState struct {
	WorkloadPolicyTargetArgs
}

// Annotate provides SDK documentation.
func (s *WorkloadPolicyTargetState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "Human-friendly name for this target.")
	a.Describe(&s.PolicyId, "Workload policy ID this target is attached to.")
	a.Describe(&s.ClusterIds, "Cluster IDs where this target applies.")
	a.Describe(&s.Description, "Free-form description of the target.")
	a.Describe(&s.Priority, "Evaluation priority; higher values take precedence when targets overlap.")
	a.Describe(&s.Enabled, "Enable or disable this target.")
	a.Describe(&s.WorkloadNames, "Explicit list of workload names to include.")
	a.Describe(&s.NodeGroupNames, "Restrict matching to specific node groups by name.")
	a.Describe(&s.KindFilter, "Restrict matching to specific Kubernetes kinds (e.g. Deployment, Pod).")
	a.Describe(&s.NamePattern, "Regex to match workload names.")
	a.Describe(&s.NamespaceSelector, "Select namespaces by labels.")
	a.Describe(&s.WorkloadSelector, "Select workloads by labels.")
}

// WorkloadPolicyTarget is the resource implementation.
type WorkloadPolicyTarget struct{}

// ---------- CRUD ----------

func (w *WorkloadPolicyTarget) Create(ctx context.Context, req infer.CreateRequest[WorkloadPolicyTargetArgs]) (infer.CreateResponse[WorkloadPolicyTargetState], error) {
	if req.DryRun {
		return infer.CreateResponse[WorkloadPolicyTargetState]{Output: WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.CreateResponse[WorkloadPolicyTargetState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.CreateWorkloadPolicyTarget(ctx, connect.NewRequest(targetArgsToCreateRequest(cs.TeamID, req.Inputs)))
	if err != nil {
		return infer.CreateResponse[WorkloadPolicyTargetState]{}, fmt.Errorf("CreateWorkloadPolicyTarget: %w", err)
	}
	if resp.Msg.Target == nil {
		return infer.CreateResponse[WorkloadPolicyTargetState]{}, fmt.Errorf("CreateWorkloadPolicyTarget: empty response from server")
	}

	return infer.CreateResponse[WorkloadPolicyTargetState]{
		ID:     resp.Msg.Target.TargetId,
		Output: WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: targetProtoToArgs(resp.Msg.Target)},
	}, nil
}

func (w *WorkloadPolicyTarget) Read(ctx context.Context, req infer.ReadRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]) (infer.ReadResponse[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.ReadResponse[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.GetWorkloadPolicyTarget(ctx, connect.NewRequest(&apiv1.GetWorkloadPolicyTargetRequest{
		TeamId:   cs.TeamID,
		TargetId: req.ID,
	}))
	if err != nil {
		return infer.ReadResponse[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetWorkloadPolicyTarget: %w", err)
	}
	if resp.Msg.Target == nil {
		return infer.ReadResponse[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetWorkloadPolicyTarget: target not found")
	}

	updatedArgs := targetProtoToArgs(resp.Msg.Target)
	return infer.ReadResponse[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID:     req.ID,
		Inputs: updatedArgs,
		State:  WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: updatedArgs},
	}, nil
}

func (w *WorkloadPolicyTarget) Update(ctx context.Context, req infer.UpdateRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]) (infer.UpdateResponse[WorkloadPolicyTargetState], error) {
	if req.DryRun {
		return infer.UpdateResponse[WorkloadPolicyTargetState]{Output: WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.UpdateResponse[WorkloadPolicyTargetState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.UpdateWorkloadPolicyTarget(ctx, connect.NewRequest(targetArgsToUpdateRequest(cs.TeamID, req.ID, req.Inputs)))
	if err != nil {
		return infer.UpdateResponse[WorkloadPolicyTargetState]{}, fmt.Errorf("UpdateWorkloadPolicyTarget: %w", err)
	}
	if resp.Msg.Target == nil {
		return infer.UpdateResponse[WorkloadPolicyTargetState]{}, fmt.Errorf("UpdateWorkloadPolicyTarget: empty response from server")
	}

	return infer.UpdateResponse[WorkloadPolicyTargetState]{
		Output: WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: targetProtoToArgs(resp.Msg.Target)},
	}, nil
}

func (w *WorkloadPolicyTarget) Delete(ctx context.Context, req infer.DeleteRequest[WorkloadPolicyTargetState]) (infer.DeleteResponse, error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.DeleteResponse{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	_, err := cs.RecommendationClient.DeleteWorkloadPolicyTarget(ctx, connect.NewRequest(&apiv1.DeleteWorkloadPolicyTargetRequest{
		TeamId:    cs.TeamID,
		TargetIds: []string{req.ID},
	}))
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("DeleteWorkloadPolicyTarget: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

// ---------- proto conversion helpers ----------

func targetArgsToCreateRequest(teamID string, a WorkloadPolicyTargetArgs) *apiv1.CreateWorkloadPolicyTargetRequest {
	r := &apiv1.CreateWorkloadPolicyTargetRequest{
		TeamId:         teamID,
		PolicyId:       a.PolicyId,
		Name:           a.Name,
		Priority:       int32(a.Priority),
		Enabled:        a.Enabled,
		ClusterIds:     a.ClusterIds,
		WorkloadNames:  a.WorkloadNames,
		NodeGroupNames: a.NodeGroupNames,
		KindFilter:     kindFilterToProto(a.KindFilter),
		NamePattern:    namePatternToProto(a.NamePattern),
		NamespaceSelector: labelSelectorToProto(a.NamespaceSelector),
		WorkloadSelector:  labelSelectorToProto(a.WorkloadSelector),
	}
	if a.Description != nil {
		r.Description = *a.Description
	}
	return r
}

func targetArgsToUpdateRequest(teamID, targetID string, a WorkloadPolicyTargetArgs) *apiv1.UpdateWorkloadPolicyTargetRequest {
	policyID := a.PolicyId
	r := &apiv1.UpdateWorkloadPolicyTargetRequest{
		TeamId:         teamID,
		TargetId:       targetID,
		PolicyId:       &policyID,
		Name:           a.Name,
		Priority:       int32(a.Priority),
		Enabled:        a.Enabled,
		ClusterIds:     a.ClusterIds,
		WorkloadNames:  a.WorkloadNames,
		NodeGroupNames: a.NodeGroupNames,
		KindFilter:     kindFilterToProto(a.KindFilter),
		NamePattern:    namePatternToProto(a.NamePattern),
		NamespaceSelector: labelSelectorToProto(a.NamespaceSelector),
		WorkloadSelector:  labelSelectorToProto(a.WorkloadSelector),
	}
	if a.Description != nil {
		r.Description = *a.Description
	}
	return r
}

func targetProtoToArgs(t *apiv1.WorkloadPolicyTarget) WorkloadPolicyTargetArgs {
	a := WorkloadPolicyTargetArgs{
		Name:              t.Name,
		PolicyId:          t.PolicyId,
		ClusterIds:        t.ClusterIds,
		Priority:          int(t.Priority),
		Enabled:           t.Enabled,
		WorkloadNames:     t.WorkloadNames,
		NodeGroupNames:    t.NodeGroupNames,
		KindFilter:        kindFilterFromProto(t.KindFilter),
		NamePattern:       namePatternFromProto(t.NamePattern),
		NamespaceSelector: labelSelectorFromProto(t.NamespaceSelector),
		WorkloadSelector:  labelSelectorFromProto(t.WorkloadSelector),
	}
	if t.Description != "" {
		a.Description = &t.Description
	}
	return a
}

func kindFilterToProto(kinds []string) []commonv1.K8SObjectKind {
	result := make([]commonv1.K8SObjectKind, 0, len(kinds))
	for _, k := range kinds {
		switch k {
		case "Pod":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_POD)
		case "Job":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_JOB)
		case "Deployment":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_DEPLOYMENT)
		case "StatefulSet":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_STATEFUL_SET)
		case "DaemonSet":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_DAEMON_SET)
		case "ReplicaSet":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_REPLICA_SET)
		case "CronJob":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_CRON_JOB)
		case "ReplicationController":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_REPLICATION_CONTROLLER)
		case "Rollout":
			result = append(result, commonv1.K8SObjectKind_K8S_OBJECT_KIND_ARGO_ROLLOUT)
		}
	}
	return result
}

func kindFilterFromProto(kinds []commonv1.K8SObjectKind) []string {
	if len(kinds) == 0 {
		return nil
	}
	result := make([]string, 0, len(kinds))
	for _, k := range kinds {
		switch k {
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_POD:
			result = append(result, "Pod")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_JOB:
			result = append(result, "Job")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_DEPLOYMENT:
			result = append(result, "Deployment")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_STATEFUL_SET:
			result = append(result, "StatefulSet")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_DAEMON_SET:
			result = append(result, "DaemonSet")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_REPLICA_SET:
			result = append(result, "ReplicaSet")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_CRON_JOB:
			result = append(result, "CronJob")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_REPLICATION_CONTROLLER:
			result = append(result, "ReplicationController")
		case commonv1.K8SObjectKind_K8S_OBJECT_KIND_ARGO_ROLLOUT:
			result = append(result, "Rollout")
		}
	}
	return result
}

func namePatternToProto(n *NamePatternArgs) *commonv1.RegexPattern {
	if n == nil {
		return nil
	}
	p := &commonv1.RegexPattern{Pattern: n.Pattern}
	if n.Flags != nil {
		p.Flags = *n.Flags
	}
	return p
}

func namePatternFromProto(p *commonv1.RegexPattern) *NamePatternArgs {
	if p == nil {
		return nil
	}
	n := &NamePatternArgs{Pattern: p.Pattern}
	if p.Flags != "" {
		n.Flags = &p.Flags
	}
	return n
}

func labelSelectorToProto(s *LabelSelectorArgs) *commonv1.LabelSelector {
	if s == nil {
		return nil
	}
	ls := &commonv1.LabelSelector{
		MatchLabels: s.MatchLabels,
	}
	for _, expr := range s.MatchExpressions {
		ls.MatchExpressions = append(ls.MatchExpressions, &commonv1.LabelSelectorRequirement{
			Key:      expr.Key,
			Operator: labelSelectorOperatorToProto(expr.Operator),
			Values:   expr.Values,
		})
	}
	return ls
}

func labelSelectorFromProto(ls *commonv1.LabelSelector) *LabelSelectorArgs {
	if ls == nil {
		return nil
	}
	s := &LabelSelectorArgs{
		MatchLabels: ls.MatchLabels,
	}
	for _, expr := range ls.MatchExpressions {
		s.MatchExpressions = append(s.MatchExpressions, LabelSelectorRequirementArgs{
			Key:      expr.Key,
			Operator: labelSelectorOperatorFromProto(expr.Operator),
			Values:   expr.Values,
		})
	}
	return s
}

func labelSelectorOperatorToProto(op string) commonv1.LabelSelectorOperator {
	switch op {
	case "In":
		return commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_IN
	case "NotIn":
		return commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_NOT_IN
	case "Exists":
		return commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_EXISTS
	case "DoesNotExist":
		return commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_DOES_NOT_EXIST
	default:
		return commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_UNSPECIFIED
	}
}

func labelSelectorOperatorFromProto(op commonv1.LabelSelectorOperator) string {
	switch op {
	case commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_IN:
		return "In"
	case commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_NOT_IN:
		return "NotIn"
	case commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_EXISTS:
		return "Exists"
	case commonv1.LabelSelectorOperator_LABEL_SELECTOR_OPERATOR_DOES_NOT_EXIST:
		return "DoesNotExist"
	default:
		return ""
	}
}
