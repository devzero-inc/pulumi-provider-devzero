package main

import (
	"github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a Cluster named "pulumi-test-cluster".
		cluster, err := resources.NewCluster(ctx, "pulumi-test-cluster", &resources.ClusterArgs{
			Name: pulumi.String("pulumi-test-cluster"),
		})
		if err != nil {
			return err
		}

		// Create a WorkloadPolicy with CPU vertical scaling enabled.
		// CpuVerticalScaling accepts a VerticalScalingArgsPtrInput; we pass a
		// VerticalScalingArgsArgs value (the concrete input struct) converted via
		// ToVerticalScalingArgsPtrOutput().
		policy, err := resources.NewWorkloadPolicy(ctx, "cpu-scaling-policy", &resources.WorkloadPolicyArgs{
			Name:        pulumi.String("cpu-scaling-policy-go"),
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

		// Create a WorkloadPolicyTarget linking the policy to the cluster for Deployments.
		// ClusterIds references the Pulumi resource ID of the cluster.
		// PolicyId references the Pulumi resource ID of the policy.
		_, err = resources.NewWorkloadPolicyTarget(ctx, "pulumi-test-cluster-deployments-target", &resources.WorkloadPolicyTargetArgs{
			Name:        pulumi.String("pulumi-test-cluster-deployments-target"),
			Description: pulumi.StringPtr("Apply cpu-scaling-policy to all Deployments in pulumi-test-cluster"),
			PolicyId:    policy.ID(),
			ClusterIds:  pulumi.StringArray{cluster.ID()},
			KindFilter:  pulumi.StringArray{pulumi.String("Deployment")},
			Enabled:     pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		// Export the cluster ID, cluster token (marked secret in the SDK), and policy ID.
		ctx.Export("clusterId", cluster.ID())
		ctx.Export("clusterToken", cluster.Token)
		ctx.Export("policyId", policy.ID())

		return nil
	})
}
