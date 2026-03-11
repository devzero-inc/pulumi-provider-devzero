import * as pulumi from "@pulumi/pulumi";
import { resources, types } from "@devzero/pulumi-devzero";

// Create a Cluster
const cluster = new resources.Cluster("prod-cluster-typescript", {
    name: "prod-cluster-typescript",
});

// Create a WorkloadPolicy with CPU vertical scaling enabled.
// The input type for cpuVerticalScaling is inputs.resources.VerticalScalingArgsArgs,
// accessed via types.inputs.resources.VerticalScalingArgsArgs at runtime but passed inline.
const policy = new resources.WorkloadPolicy("cpu-scaling-policy-ts", {
    name: "cpu-scaling-policy-ts-v26",
    description: "Policy with CPU vertical scaling enabled",
    cpuVerticalScaling: {
        enabled: true,
        targetPercentile: 0.95,
        minRequest: 50,
        maxRequest: 4000,
        maxScaleUpPercent: 100,
        maxScaleDownPercent: 25,
        overheadMultiplier: 1.1,
        limitsAdjustmentEnabled: true,
        limitMultiplier: 1.5,
    },
});

// Create a WorkloadPolicyTarget linking the policy to the cluster for Deployments.
// clusterIds references the Pulumi resource ID of the cluster (its URN-based ID).
// policyId references the Pulumi resource ID of the policy.
const target = new resources.WorkloadPolicyTarget("prod-cluster-typescript-deployments-target", {
    name: "prod-cluster-typescript-deployments-target",
    description: "Apply cpu-scaling-policy to all Deployments in prod-cluster-typescript",
    policyId: policy.id,
    clusterIds: [cluster.id],
    kindFilter: ["Deployment"],
    enabled: true,
});

// Exports
export const clusterId = cluster.id;
export const clusterToken = pulumi.secret(cluster.token);
export const policyId = policy.id;
export const targetId = target.id;
