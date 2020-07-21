#!/bin/bash

fail() {
  echo "$1"
  exit 1
}

# Create Storage Class
kubectl apply -f openstorage-e2e/storageclass.yaml || fail "failed to apply storageclass yaml"
kubectl get storageclass openstorage-sc || fail "failed to create storageclass"

# Create PVC
kubectl apply -f openstorage-e2e/pvc.yaml || fail "failed to apply PVC yaml"
kubectl get pvc openstorage-pvc || fail "failed to create pvc"

# Use PVC with MySQL deployment
kubectl apply -f openstorage-e2e/mysql.yaml || fail "failed to apply deployment yaml"
kubectl get deployment mysql || fail "failed to create deployment"
