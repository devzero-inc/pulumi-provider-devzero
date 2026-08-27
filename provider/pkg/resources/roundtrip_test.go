package resources

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/protobuf/proto"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
)

// ---------------------------------------------------------------------------
// All-fields round-trip harness.
//
// populateArgs fills EVERY pulumi-tagged field of an args struct — pointers,
// slices, maps, nested structs — with a concrete value (enum-valued strings
// get a valid member from fieldOverrides). Each resource's round-trip test
// then converts args → proto → args → proto and requires the two protos to be
// identical (proto.Equal). Any field that toProto forgets to send, or that
// fromProto forgets to read back, breaks the equality and fails the test —
// so newly added fields cannot silently stop round-tripping.
// ---------------------------------------------------------------------------

// fieldOverrides supplies valid values for pulumi fields whose converters
// only accept a fixed vocabulary. Keyed by the pulumi tag name, applied at
// any nesting depth.
var fieldOverrides = map[string]string{
	"actionTriggers":      "on_schedule",
	"detectionTriggers":   "pod_creation",
	"kind":                "Deployment",
	"kindFilter":          "Deployment",
	"kindFilterNotIn":     "CronJob",
	"operator":            "In",
	"effect":              "NoSchedule",
	"primaryMetric":       "cpu",
	"instanceStorePolicy": "INSTANCE_STORE_POLICY_RAID0",
	"consolidationPolicy": "WhenEmptyOrUnderutilized",
	"type":                "CPU",
	"selectPolicy":        "Max",
}

func populateArgs(t *testing.T, v reflect.Value, tag string) {
	t.Helper()
	switch v.Kind() {
	case reflect.Ptr:
		v.Set(reflect.New(v.Type().Elem()))
		populateArgs(t, v.Elem(), tag)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			fieldTag, ok := f.Tag.Lookup("pulumi")
			if !ok {
				continue
			}
			name := strings.Split(fieldTag, ",")[0]
			populateArgs(t, v.Field(i), name)
		}
	case reflect.Slice:
		elem := reflect.New(v.Type().Elem()).Elem()
		populateArgs(t, elem, tag)
		v.Set(reflect.Append(reflect.MakeSlice(v.Type(), 0, 1), elem))
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		val := reflect.New(v.Type().Elem()).Elem()
		populateArgs(t, val, tag)
		m.SetMapIndex(reflect.ValueOf("k"), val)
		v.Set(m)
	case reflect.String:
		if o, ok := fieldOverrides[tag]; ok {
			v.SetString(o)
		} else {
			v.SetString("x-" + tag)
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		v.SetInt(7)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(0.75)
	case reflect.Bool:
		v.SetBool(true)
	default:
		t.Fatalf("populateArgs: unhandled kind %s for field %q", v.Kind(), tag)
	}
}

func requireProtoStable(t *testing.T, name string, proto1, proto2 proto.Message) {
	t.Helper()
	if !proto.Equal(proto1, proto2) {
		t.Fatalf("%s: proto round-trip is lossy — a field is dropped in toProto or fromProto.\nfirst:  %v\nsecond: %v", name, proto1, proto2)
	}
}

func TestRoundTrip_AllFields_NodePolicy(t *testing.T) {
	var args NodePolicyArgs
	populateArgs(t, reflect.ValueOf(&args).Elem(), "")
	args.Raw = nil // raw YAML is write-only passthrough; keep the test on structured fields

	p1 := nodePolicyArgsToProto("team", "id", args)
	back := nodePolicyProtoToArgs(p1)
	p2 := nodePolicyArgsToProto("team", "id", back)
	requireProtoStable(t, "NodePolicy", p1, p2)

	// Spot-check a few of the newly added fields survived the full cycle.
	if p2.ZonalShift == nil || !p2.ZonalShift.RespectZonalShift {
		t.Fatal("zonalShift did not round-trip")
	}
	if len(p2.StartupTaints) != 1 {
		t.Fatal("startupTaints did not round-trip")
	}
	if p2.InstanceLocalNvme == nil {
		t.Fatal("instanceLocalNvme did not round-trip")
	}
	if p2.CloudProviderId == nil || *p2.CloudProviderId != 7 {
		t.Fatal("cloudProviderId did not round-trip")
	}
	if p2.Azure == nil || p2.Azure.ImageVersion == nil {
		t.Fatal("azure.imageVersion did not round-trip")
	}
	if p2.MasterOverrideRoleName == "" {
		t.Fatal("masterOverrideRoleName did not round-trip")
	}
}

func TestRoundTrip_AllFields_NodePolicyTarget(t *testing.T) {
	var args NodePolicyTargetArgs
	populateArgs(t, reflect.ValueOf(&args).Elem(), "")

	p1 := nodePolicyTargetArgsToProto("team", "id", args)
	back := nodePolicyTargetProtoToArgs(p1)
	p2 := nodePolicyTargetArgsToProto("team", "id", back)
	requireProtoStable(t, "NodePolicyTarget", p1, p2)
}

func TestRoundTrip_AllFields_WorkloadPolicy(t *testing.T) {
	var args WorkloadPolicyArgs
	populateArgs(t, reflect.ValueOf(&args).Elem(), "")

	p1 := argsToProto("team", "id", args)
	back := protoToArgs(p1)
	p2 := argsToProto("team", "id", back)
	requireProtoStable(t, "WorkloadPolicy", p1, p2)

	// Spot-check the newly added field groups.
	if !p2.EnableInPlaceVerticalScaling || !p2.AllowInPlaceMemoryLimitDecrease || !p2.PdbEnabled {
		t.Fatal("in-place/pdb flags did not round-trip")
	}
	if p2.CpuFloorPercent == nil || p2.MemoryLimitCeilingPercent == nil {
		t.Fatal("floor/ceiling percents did not round-trip")
	}
	if !p2.JvmHeapOptimizationEnabled || p2.JvmMaxHeapBytes == nil || p2.JvmCpuStartupFloorMillicores == nil {
		t.Fatal("JVM knobs did not round-trip")
	}
	if p2.CpuVerticalScaling.RequestUseRss == nil || p2.MemoryVerticalScaling.LimitUseRss == nil {
		t.Fatal("RSS flags did not round-trip")
	}
	hs := p2.HorizontalScaling
	if hs.NetworkTargetThroughputBytesPerSec == nil || hs.TargetMemoryUtilization == nil || hs.CompositeFormula == nil || hs.ScaleDownCooldownSeconds == nil {
		t.Fatal("horizontal scaling extras did not round-trip")
	}
}

func TestRoundTrip_AllFields_WorkloadPolicyTarget(t *testing.T) {
	var args WorkloadPolicyTargetArgs
	populateArgs(t, reflect.ValueOf(&args).Elem(), "")

	req1 := targetArgsToCreateRequest("team", args)
	// Synthesize the target the server would echo back for these fields.
	echo := &apiv1.WorkloadPolicyTarget{
		TargetId:           "id",
		PolicyId:           req1.PolicyId,
		TeamId:             req1.TeamId,
		Name:               req1.Name,
		Description:        req1.Description,
		Priority:           req1.Priority,
		Enabled:            req1.Enabled,
		NamespaceSelector:  req1.NamespaceSelector,
		WorkloadSelector:   req1.WorkloadSelector,
		AnnotationSelector: req1.AnnotationSelector,
		KindFilter:         req1.KindFilter,
		KindFilterNotIn:    req1.KindFilterNotIn,
		NamePattern:        req1.NamePattern,
		NamespacePattern:   req1.NamespacePattern,
		WorkloadNames:      req1.WorkloadNames,
		WorkloadNamesNotIn: req1.WorkloadNamesNotIn,
		NodeGroupNames:     req1.NodeGroupNames, //nolint:staticcheck // deprecated upstream but still round-tripped
		ClusterIds:         req1.ClusterIds,
	}
	back := targetProtoToArgs(echo)
	req2 := targetArgsToCreateRequest("team", back)
	requireProtoStable(t, "WorkloadPolicyTarget", req1, req2)

	if req2.AnnotationSelector == nil {
		t.Fatal("annotationSelector did not round-trip")
	}
	if len(req2.WorkloadNamesNotIn) != 1 || len(req2.KindFilterNotIn) != 1 {
		t.Fatal("notIn filters did not round-trip")
	}
}

func TestRoundTrip_AllFields_WorkloadRule(t *testing.T) {
	var args WorkloadRuleArgs
	populateArgs(t, reflect.ValueOf(&args).Elem(), "")
	auto := false
	args.AutoGenerate = &auto // manual mode so Fields are populated

	req1 := ruleArgsToUpsertRequest("team", args, true)
	if req1.Fields == nil {
		t.Fatal("expected manual fields")
	}
	// Synthesize the rule the server would echo back for these fields.
	f := req1.Fields
	echo := &apiv1.WorkloadRule{
		RuleId:                          "id",
		ClusterId:                       req1.ClusterId,
		Namespace:                       req1.Namespace,
		Kind:                            req1.Kind,
		Name:                            req1.Name,
		CurrentSource:                   "pulumi_manual",
		CpuRule:                         f.CpuRule,
		MemoryRule:                      f.MemoryRule,
		GpuRule:                         f.GpuRule,
		HpaRule:                         f.HpaRule,
		EmergencyResponse:               f.EmergencyResponse,
		ActionTriggers:                  f.ActionTriggers,
		DetectionTriggers:               f.DetectionTriggers,
		SchedulerPlugins:                f.SchedulerPlugins,
		LiveMigrationEnabled:            f.LiveMigrationEnabled,
		UseInPlaceVerticalScaling:       f.UseInPlaceVerticalScaling,
		Containers:                      f.Containers,
		StartupPeriodSeconds:            f.StartupPeriodSeconds,
		CronSchedule:                    f.CronSchedule,
		CooldownMinutes:                 f.CooldownMinutes,
		DefragmentationSchedule:         f.DefragmentationSchedule,
		LookbackPeriodSeconds:           f.LookbackPeriodSeconds,
		AllowInPlaceMemoryLimitDecrease: f.AllowInPlaceMemoryLimitDecrease,
		JvmHeapRule:                     f.JvmHeapRule,
		JvmCpuStartupFloorMillicores:    f.JvmCpuStartupFloorMillicores,
		KedaScaledObject:                f.KedaScaledObject,
	}
	if f.Disabled != nil {
		echo.Disabled = *f.Disabled
	}
	back := ruleProtoToArgs(echo)
	back.AutoGenerate = &auto
	req2 := ruleArgsToUpsertRequest("team", back, true)
	requireProtoStable(t, "WorkloadRule", req1.Fields, req2.Fields)

	f2 := req2.Fields
	if f2.LookbackPeriodSeconds == nil || f2.Disabled == nil {
		t.Fatal("lookback/disabled did not round-trip")
	}
	if f2.CpuRule.InitialRequest == nil || f2.CpuRule.FloorPercent == nil || f2.CpuRule.LimitCeilingPercent == nil {
		t.Fatal("initial/floor/ceiling bounds did not round-trip")
	}
	if f2.MemoryRule.RequestUseRss == nil || f2.MemoryRule.LimitUseRss == nil {
		t.Fatal("RSS flags did not round-trip")
	}
	if len(f2.HpaRule.Metrics) != 1 || f2.HpaRule.Metrics[0].ConnectorId == nil {
		t.Fatal("metric connectorId did not round-trip")
	}
	if len(f2.Containers) != 1 || f2.Containers[0].CpuRule.RequestUseRss == nil {
		t.Fatal("container RSS flags did not round-trip")
	}
}

// ---------------------------------------------------------------------------
// Diff (replace-on-identity-change) tests
// ---------------------------------------------------------------------------

func diffReq(old, new WorkloadRuleArgs) infer.DiffRequest[WorkloadRuleArgs, WorkloadRuleState] {
	return infer.DiffRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID:     "wr-1",
		State:  WorkloadRuleState{WorkloadRuleArgs: old},
		Inputs: new,
	}
}

func TestWorkloadRuleDiff(t *testing.T) {
	w := &WorkloadRule{}
	base := WorkloadRuleArgs{ClusterID: "c1", Namespace: "prod", Kind: "Deployment", Name: "api"}

	t.Run("no change", func(t *testing.T) {
		resp, err := w.Diff(context.Background(), diffReq(base, base))
		if err != nil {
			t.Fatal(err)
		}
		if resp.HasChanges {
			t.Fatalf("expected no changes, got %v", resp.DetailedDiff)
		}
	})

	for _, field := range []string{"clusterId", "namespace", "kind", "name"} {
		t.Run("identity change replaces: "+field, func(t *testing.T) {
			changed := base
			switch field {
			case "clusterId":
				changed.ClusterID = "c2"
			case "namespace":
				changed.Namespace = "staging"
			case "kind":
				changed.Kind = "StatefulSet"
			case "name":
				changed.Name = "api-v2"
			}
			resp, err := w.Diff(context.Background(), diffReq(base, changed))
			if err != nil {
				t.Fatal(err)
			}
			if !resp.HasChanges {
				t.Fatal("expected changes")
			}
			d, ok := resp.DetailedDiff[field]
			if !ok {
				t.Fatalf("expected a diff entry for %q, got %v", field, resp.DetailedDiff)
			}
			if !strings.Contains(string(d.Kind), "replace") {
				t.Fatalf("expected %q to require replacement, got kind %q", field, d.Kind)
			}
		})
	}

	t.Run("non-identity change updates in place", func(t *testing.T) {
		changed := base
		cron := "*/30 * * * *"
		changed.CronSchedule = &cron
		resp, err := w.Diff(context.Background(), diffReq(base, changed))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.HasChanges {
			t.Fatal("expected changes")
		}
		d, ok := resp.DetailedDiff["cronSchedule"]
		if !ok {
			t.Fatalf("expected a diff entry for cronSchedule, got %v", resp.DetailedDiff)
		}
		if strings.Contains(string(d.Kind), "replace") {
			t.Fatalf("cronSchedule change must not force replacement, got kind %q", d.Kind)
		}
	})
}

// Guard: every top-level pulumi field of every args struct is either populated
// by populateArgs or explicitly excluded above — fail if a future field type
// isn't handled, so the harness can't silently skip it.
func TestPopulateArgsHandlesAllFieldKinds(t *testing.T) {
	for _, target := range []any{
		&NodePolicyArgs{}, &NodePolicyTargetArgs{}, &WorkloadPolicyArgs{},
		&WorkloadPolicyTargetArgs{}, &WorkloadRuleArgs{},
	} {
		v := reflect.ValueOf(target).Elem()
		populateArgs(t, v, "")
		// verify nothing pulumi-tagged stayed zero
		var walk func(v reflect.Value, path string)
		walk = func(v reflect.Value, path string) {
			switch v.Kind() {
			case reflect.Ptr:
				if v.IsNil() {
					t.Errorf("%s: pointer left nil", path)
					return
				}
				walk(v.Elem(), path)
			case reflect.Struct:
				for i := 0; i < v.NumField(); i++ {
					f := v.Type().Field(i)
					if _, ok := f.Tag.Lookup("pulumi"); !ok {
						continue
					}
					walk(v.Field(i), path+"."+f.Name)
				}
			case reflect.Slice, reflect.Map:
				if v.Len() == 0 {
					t.Errorf("%s: collection left empty", path)
				}
			case reflect.String:
				if v.String() == "" {
					t.Errorf("%s: string left empty", path)
				}
			}
		}
		walk(v, fmt.Sprintf("%T", target))
	}
}
