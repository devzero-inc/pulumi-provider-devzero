package main

import (
	"github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Look up an existing cluster that was registered manually (not created by Pulumi).
		// This is the primary use case: find a cluster by name to get its ID,
		// so you can attach policies or inject the ID into values.yaml / a secret.
		existing, err := resources.GetClusterIdByName(ctx, &resources.GetClusterIdByNameArgs{
			Name: "pulumi-test-cluster",
		})
		if err != nil {
			return err
		}

		// Create a WorkloadPolicy.
		policy, err := resources.NewWorkloadPolicy(ctx, "cpu-scaling-policy", &resources.WorkloadPolicyArgs{
			Name:        pulumi.String("cpu-scaling-policy-go-v2"),
			Description: pulumi.StringPtr("Policy with CPU vertical scaling enabled"),
			CpuVerticalScaling: resources.VerticalScalingArgsArgs{
				Enabled:                 pulumi.BoolPtr(true),
				TargetPercentile:        pulumi.Float64Ptr(0.95),
				MinRequest:              pulumi.IntPtr(50),
				MaxRequest:              pulumi.IntPtr(4000),
				MaxScaleUpPercent:       pulumi.Float64Ptr(100),
				MaxScaleDownPercent:     pulumi.Float64Ptr(25),
				OverheadMultiplier:      pulumi.Float64Ptr(1.1),
				LimitsAdjustmentEnabled: pulumi.BoolPtr(true),
				LimitMultiplier:         pulumi.Float64Ptr(1.5),
			}.ToVerticalScalingArgsPtrOutput(),
		})
		if err != nil {
			return err
		}

		// Attach the policy to the existing (manually registered) cluster using its looked-up ID.
		_, err = resources.NewWorkloadPolicyTarget(ctx, "existing-cluster-deployments-target", &resources.WorkloadPolicyTargetArgs{
			Name:        pulumi.String("existing-cluster-deployments-target"),
			Description: pulumi.StringPtr("Apply cpu-scaling-policy-go-v2 to all Deployments in the existing cluster"),
			PolicyId:    policy.ID(),
			ClusterIds:  pulumi.StringArray{pulumi.String(existing.ClusterId)},
			KindFilter:  pulumi.StringArray{pulumi.String("Deployment")},
			Enabled:     pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		// Export the looked-up cluster ID and policy ID.
		ctx.Export("clusterId", pulumi.String(existing.ClusterId))
		ctx.Export("policyId", policy.ID())

		return nil
	})
}