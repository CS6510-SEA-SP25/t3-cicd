#!/bin/bash

### All done in namespace default
# # Debugging mode
set -x

kubectl apply -f k8s/rbac.yaml

kubectl apply -f https://raw.githubusercontent.com/mysql/mysql-operator/trunk/deploy/deploy-crds.yaml

kubectl apply -f https://raw.githubusercontent.com/mysql/mysql-operator/trunk/deploy/deploy-operator.yaml

helm install mycluster mysql-operator/mysql-innodbcluster \
        --namespace default \
        --set credentials.root.user='root' \
        --set credentials.root.password='root' \
        --set credentials.root.host='%' \
        --set serverInstances=3 \
        --set routerInstances=1 \
        --set tls.useSelfSigned=true

# kubectl patch statefulset mycluster -n default --type merge --patch-file=mysql/mysql-statefulset-patch.yaml

echo "Kubernetes setup complete!"
