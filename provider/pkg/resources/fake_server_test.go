package resources

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	apiv1connect "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1/apiv1connect"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// ---------------------------------------------------------------------------
// Fake backend: an in-memory connect server mimicking the dakr semantics that
// matter to CRUD correctness (upsert keys, NotFound codes, node-policy-target
// invariants, fields.disabled rejection on rule update, virtual
// source=="cluster" policies mixed into ListNodePolicies).
// ---------------------------------------------------------------------------

type fakeBackend struct {
	apiv1connect.UnimplementedK8SRecommendationServiceHandler
	apiv1connect.UnimplementedK8SServiceHandler
	apiv1connect.UnimplementedClusterMutationServiceHandler

	mu      sync.Mutex
	nextID  int
	nodePol map[string]*apiv1.NodePolicy
	nodeTgt map[string]*apiv1.NodePolicyTarget
	wp      map[string]*apiv1.WorkloadRecommendationPolicy
	wpt     map[string]*apiv1.WorkloadPolicyTarget
	rules   map[string]*apiv1.WorkloadRule
}

func newFakeBackend() *fakeBackend {
	return &fakeBackend{
		nodePol: map[string]*apiv1.NodePolicy{},
		nodeTgt: map[string]*apiv1.NodePolicyTarget{},
		wp:      map[string]*apiv1.WorkloadRecommendationPolicy{},
		wpt:     map[string]*apiv1.WorkloadPolicyTarget{},
		rules:   map[string]*apiv1.WorkloadRule{},
	}
}

func (f *fakeBackend) id(prefix string) string {
	f.nextID++
	return fmt.Sprintf("%s-%04d", prefix, f.nextID)
}

func (f *fakeBackend) CreateNodePolicies(_ context.Context, req *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*apiv1.NodePolicy, 0, len(req.Msg.Policies))
	for _, p := range req.Msg.Policies {
		if p.Id == "" {
			p.Id = f.id("np")
		}
		p.TeamId = req.Msg.TeamId
		f.nodePol[p.Id] = p
		out = append(out, p)
	}
	return connect.NewResponse(&apiv1.CreateNodePoliciesResponse{Policies: out}), nil
}

func (f *fakeBackend) ListNodePolicies(_ context.Context, _ *connect.Request[apiv1.ListNodePoliciesRequest]) (*connect.Response[apiv1.ListNodePoliciesResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	resp := &apiv1.ListNodePoliciesResponse{}
	for _, p := range f.nodePol {
		resp.Policies = append(resp.Policies, p)
	}
	// dakr mixes read-only virtual policies mirrored from non-dakr Karpenter
	// resources into the list; the provider must skip them.
	resp.Policies = append(resp.Policies, &apiv1.NodePolicy{
		Id: "np-virtual-cluster", Name: "cluster-managed", Source: "cluster",
	})
	return connect.NewResponse(resp), nil
}

func (f *fakeBackend) UpdateNodePolicy(_ context.Context, req *connect.Request[apiv1.UpdateNodePolicyRequest]) (*connect.Response[apiv1.UpdateNodePolicyResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p := req.Msg.Policy
	if _, ok := f.nodePol[p.Id]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("node policy not found"))
	}
	f.nodePol[p.Id] = p
	return connect.NewResponse(&apiv1.UpdateNodePolicyResponse{Policy: p}), nil
}

func (f *fakeBackend) DeleteNodePolicy(_ context.Context, req *connect.Request[apiv1.DeleteNodePolicyRequest]) (*connect.Response[apiv1.DeleteNodePolicyResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.nodePol[req.Msg.PolicyId]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("resource not found"))
	}
	delete(f.nodePol, req.Msg.PolicyId)
	var deleted int32
	for id, t := range f.nodeTgt {
		if t.PolicyId == req.Msg.PolicyId {
			delete(f.nodeTgt, id)
			deleted++
		}
	}
	return connect.NewResponse(&apiv1.DeleteNodePolicyResponse{Success: true, DeletedTargetCount: deleted}), nil
}

func (f *fakeBackend) CreateNodePolicyTargets(_ context.Context, req *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*apiv1.NodePolicyTarget, 0, len(req.Msg.Targets))
	for _, t := range req.Msg.Targets {
		if t.TargetId != "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("target_id must be empty on create"))
		}
		if len(t.ClusterIds) > 1 {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("node policy target must target at most one cluster"))
		}
		t.TargetId = f.id("npt")
		f.nodeTgt[t.TargetId] = t
		out = append(out, t)
	}
	return connect.NewResponse(&apiv1.CreateNodePolicyTargetsResponse{Targets: out}), nil
}

func (f *fakeBackend) ListNodePolicyTargets(_ context.Context, _ *connect.Request[apiv1.ListNodePolicyTargetsRequest]) (*connect.Response[apiv1.ListNodePolicyTargetsResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	resp := &apiv1.ListNodePolicyTargetsResponse{}
	for _, t := range f.nodeTgt {
		resp.Targets = append(resp.Targets, t)
	}
	return connect.NewResponse(resp), nil
}

func (f *fakeBackend) UpdateNodePolicyTarget(_ context.Context, req *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := req.Msg.Target
	if _, ok := f.nodeTgt[t.TargetId]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("target not found"))
	}
	f.nodeTgt[t.TargetId] = t
	return connect.NewResponse(&apiv1.UpdateNodePolicyTargetResponse{Target: t}), nil
}

func (f *fakeBackend) CreateWorkloadRecommendationPolicy(_ context.Context, req *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p := req.Msg.Policy
	if p == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("policy required"))
	}
	p.PolicyId = f.id("wp")
	p.TeamId = req.Msg.TeamId
	f.wp[p.PolicyId] = p
	return connect.NewResponse(&apiv1.CreateWorkloadRecommendationPolicyResponse{Policy: p}), nil
}

func (f *fakeBackend) GetWorkloadRecommendationPolicy(_ context.Context, req *connect.Request[apiv1.GetWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.GetWorkloadRecommendationPolicyResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.wp[req.Msg.PolicyId]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workload policy not found"))
	}
	return connect.NewResponse(&apiv1.GetWorkloadRecommendationPolicyResponse{Policy: p}), nil
}

func (f *fakeBackend) DeleteWorkloadRecommendationPolicy(_ context.Context, req *connect.Request[apiv1.DeleteWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.DeleteWorkloadRecommendationPolicyResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.wp[req.Msg.PolicyId]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workload policy not found"))
	}
	delete(f.wp, req.Msg.PolicyId)
	return connect.NewResponse(&apiv1.DeleteWorkloadRecommendationPolicyResponse{Success: true}), nil
}

func (f *fakeBackend) UpsertManualWorkloadRule(_ context.Context, req *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := req.Msg.ClusterId + "/" + req.Msg.Namespace + "/" + req.Msg.Kind + "/" + req.Msg.Name
	var existing *apiv1.WorkloadRule
	for _, r := range f.rules {
		if r.ClusterId+"/"+r.Namespace+"/"+r.Kind+"/"+r.Name == key {
			existing = r
			break
		}
	}
	if existing != nil && req.Msg.Fields != nil && req.Msg.Fields.Disabled != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("fields.disabled cannot be set when updating an existing rule; use ToggleWorkloadRuleDisabled instead"))
	}
	rule := &apiv1.WorkloadRule{
		ClusterId:  req.Msg.ClusterId,
		Namespace:  req.Msg.Namespace,
		Kind:       req.Msg.Kind,
		Name:       req.Msg.Name,
		Status:     "active",
		SyncStatus: "pending",
		Generation: 1,
	}
	switch req.Msg.Source {
	case apiv1.WorkloadRuleSource_WORKLOAD_RULE_SOURCE_PULUMI_AUTO:
		rule.CurrentSource = "pulumi_auto"
	default:
		rule.CurrentSource = "pulumi_manual"
	}
	if existing != nil {
		rule.RuleId = existing.RuleId
		rule.Generation = existing.Generation + 1
		rule.Disabled = existing.Disabled
	} else {
		rule.RuleId = f.id("wr")
	}
	if fields := req.Msg.Fields; fields != nil {
		rule.CpuRule = fields.CpuRule
		rule.MemoryRule = fields.MemoryRule
		rule.GpuRule = fields.GpuRule
		rule.HpaRule = fields.HpaRule
		rule.EmergencyResponse = fields.EmergencyResponse
		rule.ActionTriggers = fields.ActionTriggers
		rule.DetectionTriggers = fields.DetectionTriggers
		rule.SchedulerPlugins = fields.SchedulerPlugins
		rule.LiveMigrationEnabled = fields.LiveMigrationEnabled
		rule.UseInPlaceVerticalScaling = fields.UseInPlaceVerticalScaling
		rule.Containers = fields.Containers
		rule.LookbackPeriodSeconds = fields.LookbackPeriodSeconds
		rule.StartupPeriodSeconds = fields.StartupPeriodSeconds
		rule.CronSchedule = fields.CronSchedule
		rule.CooldownMinutes = fields.CooldownMinutes
		rule.DefragmentationSchedule = fields.DefragmentationSchedule
		if fields.Disabled != nil {
			rule.Disabled = *fields.Disabled
		}
	}
	f.rules[rule.RuleId] = rule
	return connect.NewResponse(&apiv1.UpsertManualWorkloadRuleResponse{Rule: rule}), nil
}

func (f *fakeBackend) GetWorkloadRuleByID(_ context.Context, req *connect.Request[apiv1.GetWorkloadRuleByIDRequest]) (*connect.Response[apiv1.GetWorkloadRuleByIDResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rules[req.Msg.RuleId]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workload rule not found"))
	}
	return connect.NewResponse(&apiv1.GetWorkloadRuleByIDResponse{Rule: r}), nil
}

func (f *fakeBackend) DeleteWorkloadRule(_ context.Context, req *connect.Request[apiv1.DeleteWorkloadRuleRequest]) (*connect.Response[apiv1.DeleteWorkloadRuleResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rules[req.Msg.RuleId]; !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workload rule not found"))
	}
	delete(f.rules, req.Msg.RuleId)
	return connect.NewResponse(&apiv1.DeleteWorkloadRuleResponse{}), nil
}

func (f *fakeBackend) ToggleWorkloadRuleDisabled(_ context.Context, req *connect.Request[apiv1.ToggleWorkloadRuleDisabledRequest]) (*connect.Response[apiv1.ToggleWorkloadRuleDisabledResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rules[req.Msg.RuleId]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workload rule not found"))
	}
	r.Disabled = req.Msg.Disabled
	return connect.NewResponse(&apiv1.ToggleWorkloadRuleDisabledResponse{Rule: r}), nil
}

// ---------------------------------------------------------------------------
// Harness plumbing
// ---------------------------------------------------------------------------

func withFakeServer(t *testing.T) *fakeBackend {
	t.Helper()
	fake := newFakeBackend()
	mux := http.NewServeMux()
	mux.Handle(apiv1connect.NewK8SRecommendationServiceHandler(fake))
	mux.Handle(apiv1connect.NewK8SServiceHandler(fake))
	mux.Handle(apiv1connect.NewClusterMutationServiceHandler(fake))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	prev := clientset.Get()
	clientset.Set(&clientset.ClientSet{
		TeamID:                "team-1",
		Token:                 "test-token",
		ClusterMutationClient: apiv1connect.NewClusterMutationServiceClient(srv.Client(), srv.URL),
		K8SClient:             apiv1connect.NewK8SServiceClient(srv.Client(), srv.URL),
		RecommendationClient:  apiv1connect.NewK8SRecommendationServiceClient(srv.Client(), srv.URL),
		ClusterServiceClient:  apiv1connect.NewClusterServiceClient(srv.Client(), srv.URL),
	})
	t.Cleanup(func() { clientset.Set(prev) })
	return fake
}

// ---------------------------------------------------------------------------
// Lifecycle tests
// ---------------------------------------------------------------------------

func TestLifecycle_NodePolicy(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	n := &NodePolicy{}

	args := NodePolicyArgs{
		Name: "azure-pool",
		Azure: &AzureNodeClassSpecArgs{
			VnetSubnetId: "/subscriptions/x/subnet/y",
		},
		Taints: []TaintArgs{{Key: "dedicated", Value: "gpu", Effect: "NoSchedule"}},
	}

	created, err := n.Create(ctx, infer.CreateRequest[NodePolicyArgs]{Inputs: args})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected an ID")
	}
	if created.Output.Azure == nil || created.Output.Azure.VnetSubnetId != "/subscriptions/x/subnet/y" {
		t.Fatalf("azure block not round-tripped: %+v", created.Output.Azure)
	}
	if len(created.Output.Taints) != 1 || created.Output.Taints[0].Key != "dedicated" {
		t.Fatalf("taints not round-tripped: %+v", created.Output.Taints)
	}

	// Read refreshes and must skip the virtual source=="cluster" policy.
	read, err := n.Read(ctx, infer.ReadRequest[NodePolicyArgs, NodePolicyState]{
		ID: created.ID, Inputs: args, State: created.Output,
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if read.ID != created.ID {
		t.Fatalf("Read ID: got %q", read.ID)
	}

	// Out-of-band deletion drops from state.
	fake.mu.Lock()
	delete(fake.nodePol, created.ID)
	fake.mu.Unlock()
	gone, err := n.Read(ctx, infer.ReadRequest[NodePolicyArgs, NodePolicyState]{ID: created.ID, Inputs: args, State: created.Output})
	if err != nil {
		t.Fatalf("Read after out-of-band delete: %v", err)
	}
	if gone.ID != "" {
		t.Fatalf("expected drop-from-state, got ID %q", gone.ID)
	}

	// Recreate; Delete must hit the API and be idempotent.
	created, err = n.Create(ctx, infer.CreateRequest[NodePolicyArgs]{Inputs: args})
	if err != nil {
		t.Fatalf("re-Create: %v", err)
	}
	if _, err := n.Delete(ctx, infer.DeleteRequest[NodePolicyState]{ID: created.ID, State: created.Output}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	fake.mu.Lock()
	_, still := fake.nodePol[created.ID]
	fake.mu.Unlock()
	if still {
		t.Fatal("Delete did not remove the policy from the backend")
	}
	if _, err := n.Delete(ctx, infer.DeleteRequest[NodePolicyState]{ID: created.ID, State: created.Output}); err != nil {
		t.Fatalf("Delete (already gone): %v", err)
	}
}

func TestLifecycle_NodePolicyTarget(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	n := &NodePolicyTarget{}

	args := NodePolicyTargetArgs{
		Name:       "target-1",
		PolicyId:   "np-0001",
		ClusterIds: []string{"cluster-1"},
	}

	created, err := n.Create(ctx, infer.CreateRequest[NodePolicyTargetArgs]{Inputs: args})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected target id")
	}

	// Destroy disables (no delete RPC exists).
	if _, err := n.Delete(ctx, infer.DeleteRequest[NodePolicyTargetState]{ID: created.ID, State: created.Output}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	fake.mu.Lock()
	tgt := fake.nodeTgt[created.ID]
	fake.mu.Unlock()
	if tgt == nil {
		t.Fatal("target should still exist backend-side")
	}
	if tgt.Enabled {
		t.Fatal("Delete must disable the target")
	}
}

func TestLifecycle_WorkloadPolicy(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	w := &WorkloadPolicy{}

	tp := 0.75
	args := WorkloadPolicyArgs{
		Name: "policy-1",
		CpuVerticalScaling: &VerticalScalingArgs{
			Enabled:          true,
			TargetPercentile: &tp,
		},
	}

	created, err := w.Create(ctx, infer.CreateRequest[WorkloadPolicyArgs]{Inputs: args})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Output.CpuVerticalScaling == nil || created.Output.CpuVerticalScaling.TargetPercentile == nil ||
		*created.Output.CpuVerticalScaling.TargetPercentile != 0.75 {
		t.Fatalf("targetPercentile did not round-trip exactly: %+v", created.Output.CpuVerticalScaling)
	}

	// Out-of-band delete → drop from state.
	fake.mu.Lock()
	delete(fake.wp, created.ID)
	fake.mu.Unlock()
	gone, err := w.Read(ctx, infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]{ID: created.ID, Inputs: args, State: created.Output})
	if err != nil {
		t.Fatalf("Read after out-of-band delete: %v", err)
	}
	if gone.ID != "" {
		t.Fatalf("expected drop-from-state, got ID %q", gone.ID)
	}
}

func TestLifecycle_WorkloadRule(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	w := &WorkloadRule{}

	minReq := 10
	args := WorkloadRuleArgs{
		ClusterID: "cluster-1",
		Namespace: "prod",
		Kind:      "Deployment",
		Name:      "api",
		CpuRule:   &ResourceRuleConfigArgs{Enabled: true, MinRequest: &minReq},
	}

	created, err := w.Create(ctx, infer.CreateRequest[WorkloadRuleArgs]{Inputs: args})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected rule id")
	}

	// Update flipping `disabled` must go through the toggle RPC (the fake
	// rejects fields.disabled on update, mirroring dakr).
	disabled := true
	updatedArgs := args
	updatedArgs.Disabled = &disabled
	updated, err := w.Update(ctx, infer.UpdateRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: created.ID, Inputs: updatedArgs, State: created.Output,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Output.Disabled == nil || !*updated.Output.Disabled {
		t.Fatalf("expected disabled=true in output, got %+v", updated.Output.Disabled)
	}
	fake.mu.Lock()
	rule := fake.rules[created.ID]
	fake.mu.Unlock()
	if rule == nil || !rule.Disabled {
		t.Fatalf("expected rule disabled via toggle RPC, got %+v", rule)
	}

	// Delete + idempotent delete.
	if _, err := w.Delete(ctx, infer.DeleteRequest[WorkloadRuleState]{ID: created.ID, State: updated.Output}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := w.Delete(ctx, infer.DeleteRequest[WorkloadRuleState]{ID: created.ID, State: updated.Output}); err != nil {
		t.Fatalf("Delete (already gone): %v", err)
	}
}

// Round-trip guard for the enabled default: omitting `enabled` must create an
// ENABLED target (the API stores the proto bool verbatim).
func TestLifecycle_NodePolicyTarget_EnabledDefaultsTrue(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	n := &NodePolicyTarget{}

	created, err := n.Create(ctx, infer.CreateRequest[NodePolicyTargetArgs]{Inputs: NodePolicyTargetArgs{
		Name:       "defaults",
		PolicyId:   "np-0001",
		ClusterIds: []string{"c-1"},
		// Enabled omitted.
	}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	fake.mu.Lock()
	tgt := fake.nodeTgt[created.ID]
	fake.mu.Unlock()
	if tgt == nil || !tgt.Enabled {
		t.Fatalf("expected the target to be created enabled, got %+v", tgt)
	}
}
