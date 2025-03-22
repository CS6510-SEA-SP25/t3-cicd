#!/bin/bash

# # Debugging mode
# set -x

echo "Add the MinIO Operator Repo to Helm..."
helm repo add minio-operator https://operator.min.io
echo "---------------------"

echo "Install the Operator..."
helm install --namespace minio-operator --create-namespace operator minio-operator/operator
echo "---------------------"

echo "Verify the Operator installation..."
kubectl get all -n minio-operator
echo "---------------------"

echo "Deploy the Tenant..."
helm install --namespace minio-tenant --create-namespace --values values.yaml myminio minio-operator/tenant
echo "---------------------"

echo "Add SQL file to ConfigMap..."
kubectl create configmap mysql-init-config --from-file=backend/db/init.sql
echo "---------------------"

echo "Apply k8s..."
kubectl apply -f k8s/
echo "---------------------"

# Wait for the MinIO Console service to be available before port-forwarding
echo "Waiting for all pods in namespace minio-tenant to be running..."
# kubectl wait --for=condition=ready pods --all -n minio-tenant --timeout=600s #### Failed bc pods are not created yet

while true; do
    # Get all pod names in minio-tenant namespace
    PODS=$(kubectl get pods -n minio-tenant -o jsonpath='{.items[*].metadata.name}')

    # If pods exist and are not in the "Pending" or other non-Running states, exit the loop
    if [ -n "$PODS" ] && [ -z "$(kubectl get pods -n minio-tenant --field-selector=status.phase!=Running -o jsonpath='{.items[*].status.phase}' | tr -d '[:space:]')" ]; then
        echo "All pods are running!"
        break
    else
        echo "Waiting for all pods to be running..."
        sleep 5
    fi
done

echo "---------------------"

# echo "Expose MinIO endpoints in background..."
# kubectl port-forward -n minio-tenant svc/myminio-console 9090:9090 &
# kubectl port-forward -n minio-tenant svc/myminio-hl 9000:9000 &
# echo "---------------------"

echo "Please run 'minikube tunnel' to connect to LoadBalancer services."

echo "Kubernetes setup complete!"
