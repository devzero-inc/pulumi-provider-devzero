package resources

import (
	"context"
	"reflect"
	"testing"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
)

// ---------------------------------------------------------------------------
// Import-lifecycle harness.
//
// `pulumi import` calls Read with only an ID — Inputs and State are the
// zero value, since there is no prior Pulumi state for a resource created
// out-of-band (e.g. via the DevZero UI). These tests simulate exactly that
// entry point, then prove the three things the import ticket cares about:
//
//  1. hydration   — every writable field the backend holds comes back in Inputs.
//  2. clean plan  — feeding the imported Inputs/State back into Read is a
//     fixed point, i.e. a `pulumi preview` run immediately after import
//     would compute no diff (the default infer.Diff is a structural
//     comparison of Inputs, so a fixed point on Read is exactly what a
//     clean plan requires). WorkloadRule additionally gets a direct Diff()
//     check since it defines a custom Diff.
//  3. round-trip  — modifying one field and calling Update propagates to
//     the backend.
// ---------------------------------------------------------------------------

func TestImport_NodePolicy(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	n := &NodePolicy{}

	var seed NodePolicyArgs
	populateArgs(t, reflect.ValueOf(&seed).Elem(), "")
	seed.Raw = nil // write-only passthrough, excluded like the round-trip test

	seedProto := nodePolicyArgsToProto("team-1", "np-import", seed)
	fake.mu.Lock()
	fake.nodePol[seedProto.Id] = seedProto
	fake.mu.Unlock()

	imported, err := n.Read(ctx, infer.ReadRequest[NodePolicyArgs, NodePolicyState]{ID: seedProto.Id})
	if err != nil {
		t.Fatalf("Read (import): %v", err)
	}
	want := nodePolicyProtoToArgs(seedProto)
	if !reflect.DeepEqual(imported.Inputs, want) {
		t.Fatalf("import did not hydrate all fields:\n got:  %+v\n want: %+v", imported.Inputs, want)
	}

	again, err := n.Read(ctx, infer.ReadRequest[NodePolicyArgs, NodePolicyState]{
		ID: seedProto.Id, Inputs: imported.Inputs, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Read (post-import): %v", err)
	}
	if !reflect.DeepEqual(again.Inputs, imported.Inputs) {
		t.Fatalf("post-import plan is not clean — Read is not a fixed point:\n first:  %+v\n second: %+v", imported.Inputs, again.Inputs)
	}

	changed := imported.Inputs
	changed.Name = "renamed-after-import"
	updated, err := n.Update(ctx, infer.UpdateRequest[NodePolicyArgs, NodePolicyState]{
		ID: seedProto.Id, Inputs: changed, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Update (after import): %v", err)
	}
	if updated.Output.Name != "renamed-after-import" {
		t.Fatalf("update after import did not apply: %+v", updated.Output)
	}
	fake.mu.Lock()
	stored := fake.nodePol[seedProto.Id]
	fake.mu.Unlock()
	if stored.Name != "renamed-after-import" {
		t.Fatalf("update after import was not persisted backend-side: %+v", stored)
	}
}

func TestImport_NodePolicyTarget(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	n := &NodePolicyTarget{}

	var seed NodePolicyTargetArgs
	populateArgs(t, reflect.ValueOf(&seed).Elem(), "")
	seed.ClusterIds = []string{"cluster-1"} // node policy targets accept at most one cluster

	seedProto := nodePolicyTargetArgsToProto("team-1", "npt-import", seed)
	fake.mu.Lock()
	fake.nodeTgt[seedProto.TargetId] = seedProto
	fake.mu.Unlock()

	imported, err := n.Read(ctx, infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]{ID: seedProto.TargetId})
	if err != nil {
		t.Fatalf("Read (import): %v", err)
	}
	want := nodePolicyTargetProtoToArgs(seedProto)
	if !reflect.DeepEqual(imported.Inputs, want) {
		t.Fatalf("import did not hydrate all fields:\n got:  %+v\n want: %+v", imported.Inputs, want)
	}

	again, err := n.Read(ctx, infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID: seedProto.TargetId, Inputs: imported.Inputs, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Read (post-import): %v", err)
	}
	if !reflect.DeepEqual(again.Inputs, imported.Inputs) {
		t.Fatalf("post-import plan is not clean:\n first:  %+v\n second: %+v", imported.Inputs, again.Inputs)
	}

	changed := imported.Inputs
	changed.Name = "renamed-after-import"
	updated, err := n.Update(ctx, infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID: seedProto.TargetId, Inputs: changed, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Update (after import): %v", err)
	}
	if updated.Output.Name != "renamed-after-import" {
		t.Fatalf("update after import did not apply: %+v", updated.Output)
	}
	fake.mu.Lock()
	stored := fake.nodeTgt[seedProto.TargetId]
	fake.mu.Unlock()
	if stored.Name != "renamed-after-import" {
		t.Fatalf("update after import was not persisted backend-side: %+v", stored)
	}
}

func TestImport_WorkloadPolicy(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	w := &WorkloadPolicy{}

	var seed WorkloadPolicyArgs
	populateArgs(t, reflect.ValueOf(&seed).Elem(), "")

	seedProto := argsToProto("team-1", "wp-import", seed)
	fake.mu.Lock()
	fake.wp[seedProto.PolicyId] = seedProto
	fake.mu.Unlock()

	imported, err := w.Read(ctx, infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]{ID: seedProto.PolicyId})
	if err != nil {
		t.Fatalf("Read (import): %v", err)
	}
	want := protoToArgs(seedProto)
	if !reflect.DeepEqual(imported.Inputs, want) {
		t.Fatalf("import did not hydrate all fields:\n got:  %+v\n want: %+v", imported.Inputs, want)
	}

	again, err := w.Read(ctx, infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID: seedProto.PolicyId, Inputs: imported.Inputs, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Read (post-import): %v", err)
	}
	if !reflect.DeepEqual(again.Inputs, imported.Inputs) {
		t.Fatalf("post-import plan is not clean:\n first:  %+v\n second: %+v", imported.Inputs, again.Inputs)
	}

	changed := imported.Inputs
	changed.Name = "renamed-after-import"
	updated, err := w.Update(ctx, infer.UpdateRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID: seedProto.PolicyId, Inputs: changed, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Update (after import): %v", err)
	}
	if updated.Output.Name != "renamed-after-import" {
		t.Fatalf("update after import did not apply: %+v", updated.Output)
	}
	fake.mu.Lock()
	stored := fake.wp[seedProto.PolicyId]
	fake.mu.Unlock()
	if stored.Name != "renamed-after-import" {
		t.Fatalf("update after import was not persisted backend-side: %+v", stored)
	}
}

func TestImport_WorkloadPolicyTarget(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	w := &WorkloadPolicyTarget{}

	var seed WorkloadPolicyTargetArgs
	populateArgs(t, reflect.ValueOf(&seed).Elem(), "")

	req := targetArgsToCreateRequest("team-1", seed)
	seedProto := &apiv1.WorkloadPolicyTarget{
		TargetId:           "wpt-import",
		PolicyId:           req.PolicyId,
		TeamId:             req.TeamId,
		Name:               req.Name,
		Description:        req.Description,
		Priority:           req.Priority,
		Enabled:            req.Enabled,
		NamespaceSelector:  req.NamespaceSelector,
		WorkloadSelector:   req.WorkloadSelector,
		AnnotationSelector: req.AnnotationSelector,
		KindFilter:         req.KindFilter,
		KindFilterNotIn:    req.KindFilterNotIn,
		NamePattern:        req.NamePattern,
		NamespacePattern:   req.NamespacePattern,
		WorkloadNames:      req.WorkloadNames,
		WorkloadNamesNotIn: req.WorkloadNamesNotIn,
		ClusterIds:         req.ClusterIds,
	}
	seedProto.NodeGroupNames = req.NodeGroupNames //nolint:staticcheck // deprecated upstream but still round-tripped
	fake.mu.Lock()
	fake.wpt[seedProto.TargetId] = seedProto
	fake.mu.Unlock()

	imported, err := w.Read(ctx, infer.ReadRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{ID: seedProto.TargetId})
	if err != nil {
		t.Fatalf("Read (import): %v", err)
	}
	want := targetProtoToArgs(seedProto)
	if !reflect.DeepEqual(imported.Inputs, want) {
		t.Fatalf("import did not hydrate all fields:\n got:  %+v\n want: %+v", imported.Inputs, want)
	}

	again, err := w.Read(ctx, infer.ReadRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID: seedProto.TargetId, Inputs: imported.Inputs, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Read (post-import): %v", err)
	}
	if !reflect.DeepEqual(again.Inputs, imported.Inputs) {
		t.Fatalf("post-import plan is not clean:\n first:  %+v\n second: %+v", imported.Inputs, again.Inputs)
	}

	changed := imported.Inputs
	changed.Name = "renamed-after-import"
	updated, err := w.Update(ctx, infer.UpdateRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID: seedProto.TargetId, Inputs: changed, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Update (after import): %v", err)
	}
	if updated.Output.Name != "renamed-after-import" {
		t.Fatalf("update after import did not apply: %+v", updated.Output)
	}
	fake.mu.Lock()
	stored := fake.wpt[seedProto.TargetId]
	fake.mu.Unlock()
	if stored.Name != "renamed-after-import" {
		t.Fatalf("update after import was not persisted backend-side: %+v", stored)
	}
}

func TestImport_WorkloadRule(t *testing.T) {
	ctx := context.Background()
	fake := withFakeServer(t)
	w := &WorkloadRule{}

	var seed WorkloadRuleArgs
	populateArgs(t, reflect.ValueOf(&seed).Elem(), "")
	auto := false
	seed.AutoGenerate = &auto // manual mode, so Fields get populated below

	req := ruleArgsToUpsertRequest("team-1", seed, true)
	f := req.Fields
	seedProto := &apiv1.WorkloadRule{
		RuleId:                    "wr-import",
		ClusterId:                 req.ClusterId,
		Namespace:                 req.Namespace,
		Kind:                      req.Kind,
		Name:                      req.Name,
		CurrentSource:             "pulumi_manual",
		CpuRule:                   f.CpuRule,
		MemoryRule:                f.MemoryRule,
		GpuRule:                   f.GpuRule,
		HpaRule:                   f.HpaRule,
		EmergencyResponse:         f.EmergencyResponse,
		ActionTriggers:            f.ActionTriggers,
		DetectionTriggers:         f.DetectionTriggers,
		SchedulerPlugins:          f.SchedulerPlugins,
		LiveMigrationEnabled:      f.LiveMigrationEnabled,
		UseInPlaceVerticalScaling: f.UseInPlaceVerticalScaling,
		Containers:                f.Containers,
		StartupPeriodSeconds:      f.StartupPeriodSeconds,
		CronSchedule:              f.CronSchedule,
		CooldownMinutes:           f.CooldownMinutes,
		DefragmentationSchedule:   f.DefragmentationSchedule,
		LookbackPeriodSeconds:     f.LookbackPeriodSeconds,
	}
	if f.Disabled != nil {
		seedProto.Disabled = *f.Disabled
	}
	fake.mu.Lock()
	fake.rules[seedProto.RuleId] = seedProto
	fake.mu.Unlock()

	imported, err := w.Read(ctx, infer.ReadRequest[WorkloadRuleArgs, WorkloadRuleState]{ID: seedProto.RuleId})
	if err != nil {
		t.Fatalf("Read (import): %v", err)
	}
	want := ruleProtoToArgs(seedProto)
	if !reflect.DeepEqual(imported.Inputs, want) {
		t.Fatalf("import did not hydrate all fields:\n got:  %+v\n want: %+v", imported.Inputs, want)
	}

	// Clean-plan check via the resource's own Diff (WorkloadRule overrides
	// the default structural diff), not just a second Read.
	diffResp, err := w.Diff(ctx, infer.DiffRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: seedProto.RuleId, State: imported.State, Inputs: imported.Inputs,
	})
	if err != nil {
		t.Fatalf("Diff (post-import): %v", err)
	}
	if diffResp.HasChanges {
		t.Fatalf("post-import plan is not clean: %+v", diffResp.DetailedDiff)
	}

	changed := imported.Inputs
	cron := "*/15 * * * *"
	changed.CronSchedule = &cron
	updated, err := w.Update(ctx, infer.UpdateRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: seedProto.RuleId, Inputs: changed, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Update (after import): %v", err)
	}
	if updated.Output.CronSchedule == nil || *updated.Output.CronSchedule != cron {
		t.Fatalf("update after import did not apply: %+v", updated.Output)
	}
	fake.mu.Lock()
	stored := fake.rules[seedProto.RuleId]
	fake.mu.Unlock()
	if stored.CronSchedule == nil || *stored.CronSchedule != cron {
		t.Fatalf("update after import was not persisted backend-side: %+v", stored)
	}
}

func TestImport_Cluster(t *testing.T) {
	ctx := context.Background()
	c := &Cluster{}

	const clusterID = "cl-import"
	backend := map[string]*apiv1.Cluster{
		clusterID: {Id: clusterID, CustomName: "ui-created-cluster"},
	}

	k8s := &mockK8SClient{
		getClusterFn: func(_ context.Context, req *connect.Request[apiv1.GetClusterRequest]) (*connect.Response[apiv1.GetClusterResponse], error) {
			cl, ok := backend[req.Msg.ClusterId]
			if !ok {
				return nil, connect.NewError(connect.CodeNotFound, nil)
			}
			return connect.NewResponse(&apiv1.GetClusterResponse{Cluster: cl}), nil
		},
	}
	mut := &mockMutationClient{
		updateFn: func(_ context.Context, req *connect.Request[apiv1.UpdateClusterRequest]) (*connect.Response[apiv1.UpdateClusterResponse], error) {
			cl := backend[req.Msg.ClusterId]
			cl.CustomName = req.Msg.ClusterName
			return connect.NewResponse(&apiv1.UpdateClusterResponse{Cluster: cl}), nil
		},
	}
	withMockClientSet(t, mut, k8s)

	imported, err := c.Read(ctx, infer.ReadRequest[ClusterArgs, ClusterState]{ID: clusterID})
	if err != nil {
		t.Fatalf("Read (import): %v", err)
	}
	if imported.Inputs.Name != "ui-created-cluster" {
		t.Fatalf("import did not hydrate name: %+v", imported.Inputs)
	}
	if imported.State.Token != "" {
		t.Fatalf("expected empty token on a freshly imported cluster, got %q", imported.State.Token)
	}

	again, err := c.Read(ctx, infer.ReadRequest[ClusterArgs, ClusterState]{
		ID: clusterID, Inputs: imported.Inputs, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Read (post-import): %v", err)
	}
	if !reflect.DeepEqual(again.Inputs, imported.Inputs) || again.State.Token != imported.State.Token {
		t.Fatalf("post-import plan is not clean:\n first:  %+v/%q\n second: %+v/%q",
			imported.Inputs, imported.State.Token, again.Inputs, again.State.Token)
	}

	changed := imported.Inputs
	changed.Name = "renamed-after-import"
	updated, err := c.Update(ctx, infer.UpdateRequest[ClusterArgs, ClusterState]{
		ID: clusterID, Inputs: changed, State: imported.State,
	})
	if err != nil {
		t.Fatalf("Update (after import): %v", err)
	}
	if updated.Output.Name != "renamed-after-import" {
		t.Fatalf("update after import did not apply: %+v", updated.Output)
	}
	if backend[clusterID].CustomName != "renamed-after-import" {
		t.Fatalf("update after import was not persisted backend-side: %+v", backend[clusterID])
	}
}
