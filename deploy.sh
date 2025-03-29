#!/bin/bash

### All done in namespace default
# # Debugging mode
# set -x

echo "Install the Operator..."
helm install minio-operator ./minio/operator
echo "---------------------"

echo "Verify the Operator installation..."
kubectl get all
echo "---------------------"

echo "Deploy the Tenant..."
helm install --values ./minio/values.yaml myminio minio-operator/tenant
echo "---------------------"

echo "Add CA Certificate to Secret..."
kubectl create secret generic db-ca-cert --from-file=ca.pem
echo "---------------------"

echo "Apply k8s..."
kubectl apply -f k8s/
echo "---------------------"

# Wait for the MinIO Console service to be available before port-forwarding
echo "Waiting for all pods to be running..."
# kubectl wait --for=condition=ready pods --all --timeout=600s #### Failed bc pods are not created yet

while true; do
    # Get all pod names in default namespace
    PODS=$(kubectl get pods -l app=minio-pool -o jsonpath='{.items[*].metadata.name}')

    # If pods exist and are not in the "Pending" or other non-Running states, exit the loop
    if [ -n "$PODS" ] && [ -z "$(kubectl get pods -l app=minio-pool --field-selector=status.phase!=Running -o jsonpath='{.items[*].status.phase}' | tr -d '[:space:]')" ]; then
        echo "All pods are running!"
        break
    else
        echo "Waiting for all pods to be running..."
        sleep 5
    fi
done

echo "---------------------"

# echo "Expose MinIO endpoints in background..."
# kubectl port-forward svc/myminio-console 9090:9090 &
# kubectl port-forward svc/myminio-hl 9000:9000 &
# echo "---------------------"

echo "Please run 'minikube tunnel' to connect to LoadBalancer services."

echo "Kubernetes setup complete!"
