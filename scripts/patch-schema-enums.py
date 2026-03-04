#!/usr/bin/env python3
"""
Patches schema.json to add enum constraints to fixed-value string fields.
Run after gen-schema to enrich the auto-generated schema with allowed values.
"""

import json
import sys

SCHEMA_PATH = "schema.json"


def apply_patches(schema: dict) -> dict:
    resources = schema.get("resources", {})
    types = schema.get("types", {})

    # ---------- WorkloadPolicy: actionTriggers and detectionTriggers ----------
    # These are arrays of strings; we add enum to the items schema
    for rkey, rval in resources.items():
        if rkey.endswith(":WorkloadPolicy") and not rkey.endswith(":WorkloadPolicyTarget"):
            props = rval.get("inputProperties", {})
            if "actionTriggers" in props:
                props["actionTriggers"]["items"] = {
                    "type": "string",
                    "enum": ["on_detection", "on_schedule"],
                }
                print(f"  Patched {rkey}.actionTriggers")
            if "detectionTriggers" in props:
                props["detectionTriggers"]["items"] = {
                    "type": "string",
                    "enum": ["pod_creation", "pod_update", "pod_reschedule"],
                }
                print(f"  Patched {rkey}.detectionTriggers")

    # ---------- WorkloadPolicyTarget: kindFilter ----------
    for rkey, rval in resources.items():
        if rkey.endswith(":WorkloadPolicyTarget"):
            props = rval.get("inputProperties", {})
            if "kindFilter" in props:
                props["kindFilter"]["items"] = {
                    "type": "string",
                    "enum": [
                        "Pod", "Job", "Deployment", "StatefulSet", "DaemonSet",
                        "ReplicaSet", "CronJob", "ReplicationController", "Rollout",
                    ],
                }
                print(f"  Patched {rkey}.kindFilter")

    # ---------- HorizontalScalingArgs: primaryMetric ----------
    for tkey, tval in types.items():
        if tkey.endswith(":HorizontalScalingArgs"):
            props = tval.get("properties", {})
            if "primaryMetric" in props:
                props["primaryMetric"]["type"] = "string"
                props["primaryMetric"]["enum"] = [
                    "cpu", "memory", "gpu", "network_ingress", "network_egress"
                ]
                print(f"  Patched {tkey}.primaryMetric")

    # ---------- LabelSelectorRequirementArgs: operator ----------
    for tkey, tval in types.items():
        if tkey.endswith(":LabelSelectorRequirementArgs"):
            props = tval.get("properties", {})
            if "operator" in props:
                props["operator"]["type"] = "string"
                props["operator"]["enum"] = ["In", "NotIn", "Exists", "DoesNotExist"]
                print(f"  Patched {tkey}.operator")

    # ---------- TaintArgs: effect ----------
    for tkey, tval in types.items():
        if tkey.endswith(":TaintArgs"):
            props = tval.get("properties", {})
            if "effect" in props:
                props["effect"]["type"] = "string"
                props["effect"]["enum"] = ["NoSchedule", "PreferNoSchedule", "NoExecute"]
                print(f"  Patched {tkey}.effect")

    # ---------- DisruptionPolicyArgs: consolidationPolicy ----------
    for tkey, tval in types.items():
        if tkey.endswith(":DisruptionPolicyArgs"):
            props = tval.get("properties", {})
            if "consolidationPolicy" in props:
                props["consolidationPolicy"]["type"] = "string"
                props["consolidationPolicy"]["enum"] = ["WhenEmpty", "WhenUnderutilized"]
                print(f"  Patched {tkey}.consolidationPolicy")

    return schema


if __name__ == "__main__":
    with open(SCHEMA_PATH) as f:
        schema = json.load(f)

    print("Applying enum patches to schema.json...")
    schema = apply_patches(schema)

    with open(SCHEMA_PATH, "w") as f:
        json.dump(schema, f, indent=4)

    print("Done. schema.json updated with enum constraints.")
