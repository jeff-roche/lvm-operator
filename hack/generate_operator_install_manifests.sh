#!/bin/bash
set -euo pipefail

manifests_dir="manifests/operator-install"

mkdir -p ${manifests_dir}

# Create the namespace for the operator
cat <<EOF > ${manifests_dir}/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    openshift.io/cluster-monitoring: "true"
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged
  name: openshift-storage
EOF

# Create an operatorgroup manifest for the operator
cat <<EOF > ${manifests_dir}/operatorgroup.yaml
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: lvm-operator-group
  namespace: openshift-storage
spec:
  targetNamespaces:
  - openshift-storage
EOF

# Create a Subscription manifest for the operator
cat <<EOF > ${manifests_dir}/operatorsubscription.yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: lvm-operator-subscription
  namespace: openshift-storage
spec:
  channel: ${OPERATOR_CHANNEL}
  name: lvms-operator
  source: ${CATALOG_SOURCE}
  sourceNamespace: openshift-marketplace
  installPlanApproval: Automatic
EOF

oc apply -Rf manifests
