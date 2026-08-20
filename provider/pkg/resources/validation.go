package resources

import (
	"fmt"
	"sort"
	"strings"
)

// The DevZero API silently ignores unknown enum values (an unknown trigger is
// dropped, an unknown kind becomes UNSPECIFIED and matches nothing), so the
// provider validates every string-enum input up front and fails the operation
// with an actionable message instead.

var (
	validActionTriggers = map[string]bool{
		"on_schedule":  true,
		"on_detection": true,
	}
	validDetectionTriggers = map[string]bool{
		"pod_creation":   true,
		"pod_update":     true,
		"pod_reschedule": true,
	}
	validKinds = map[string]bool{
		"Pod": true, "Job": true, "Deployment": true, "StatefulSet": true,
		"DaemonSet": true, "ReplicaSet": true, "CronJob": true,
		"ReplicationController": true, "Rollout": true,
	}
	validSelectorOperators = map[string]bool{
		"In": true, "NotIn": true, "Exists": true, "DoesNotExist": true, "Gt": true, "Lt": true,
	}
	validHPAMetrics = map[string]bool{
		"cpu": true, "memory": true, "gpu": true, "network_ingress": true, "network_egress": true,
	}
)

func sortedKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func validateEnumList(field string, values []string, valid map[string]bool) error {
	for _, v := range values {
		if !valid[v] {
			return fmt.Errorf("%s: unknown value %q (valid values: %s)", field, v, sortedKeys(valid))
		}
	}
	return nil
}

func validateSelector(field string, sel *LabelSelectorArgs) error {
	if sel == nil {
		return nil
	}
	for i, expr := range sel.MatchExpressions {
		if !validSelectorOperators[expr.Operator] {
			return fmt.Errorf("%s.matchExpressions[%d].operator: unknown value %q (valid values: %s)", field, i, expr.Operator, sortedKeys(validSelectorOperators))
		}
	}
	return nil
}

func validateWorkloadPolicyArgs(a WorkloadPolicyArgs) error {
	if err := validateEnumList("actionTriggers", a.ActionTriggers, validActionTriggers); err != nil {
		return err
	}
	if err := validateEnumList("detectionTriggers", a.DetectionTriggers, validDetectionTriggers); err != nil {
		return err
	}
	if a.HorizontalScaling != nil && a.HorizontalScaling.PrimaryMetric != nil && !validHPAMetrics[*a.HorizontalScaling.PrimaryMetric] {
		return fmt.Errorf("horizontalScaling.primaryMetric: unknown value %q (valid values: %s)", *a.HorizontalScaling.PrimaryMetric, sortedKeys(validHPAMetrics))
	}
	return nil
}

func validateWorkloadPolicyTargetArgs(a WorkloadPolicyTargetArgs) error {
	if err := validateEnumList("kindFilter", a.KindFilter, validKinds); err != nil {
		return err
	}
	if err := validateEnumList("kindFilterNotIn", a.KindFilterNotIn, validKinds); err != nil {
		return err
	}
	if err := validateSelector("namespaceSelector", a.NamespaceSelector); err != nil {
		return err
	}
	if err := validateSelector("workloadSelector", a.WorkloadSelector); err != nil {
		return err
	}
	return validateSelector("annotationSelector", a.AnnotationSelector)
}

func validateWorkloadRuleArgs(a WorkloadRuleArgs) error {
	if err := validateEnumList("actionTriggers", a.ActionTriggers, validActionTriggers); err != nil {
		return err
	}
	return validateEnumList("detectionTriggers", a.DetectionTriggers, validDetectionTriggers)
}
