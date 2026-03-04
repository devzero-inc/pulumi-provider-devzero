"""
Example DevZero infrastructure using the Python SDK.

Creates:
  - A Cluster named "prod-cluster"
  - A WorkloadPolicy with CPU vertical scaling enabled
  - A WorkloadPolicyTarget linking the policy to the cluster for Deployments
"""

import pulumi
from pulumi_devzero.resources import (
    Cluster,
    ClusterArgs,
    WorkloadPolicy,
    WorkloadPolicyArgs,
    WorkloadPolicyTarget,
    WorkloadPolicyTargetArgs,
    VerticalScalingArgsArgs,
)

# Create a cluster
cluster = Cluster(
    "prod-cluster",
    args=ClusterArgs(
        name="prod-cluster",
    ),
)

# Create a workload policy with CPU vertical scaling enabled
policy = WorkloadPolicy(
    "cpu-scaling-policy",
    args=WorkloadPolicyArgs(
        name="cpu-scaling-policy",
        description="Workload policy with CPU vertical scaling for production cluster",
        cpu_vertical_scaling=VerticalScalingArgsArgs(
            enabled=True,
        ),
    ),
)

# Create a workload policy target linking the policy to the cluster for Deployments
target = WorkloadPolicyTarget(
    "prod-cluster-deployments-target",
    args=WorkloadPolicyTargetArgs(
        name="prod-cluster-deployments-target",
        policy_id=policy.id,
        cluster_ids=[cluster.id],
        kind_filter=["Deployment"],
        enabled=True,
    ),
)

# Exports
pulumi.export("cluster_id", cluster.id)
pulumi.export("cluster_token", pulumi.Output.secret(cluster.token))
pulumi.export("policy_id", policy.id)
pulumi.export("target_id", target.id)
