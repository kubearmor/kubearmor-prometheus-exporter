#!/bin/bash

NAMESPACE=kubearmor

if [ ! -z $1 ]; then
    NAMESPACE=$1
else
    echo "Default Namespace: $NAMESPACE"
fi

KUBEARMOR_EXPORTER=$(kubectl get pods -n $NAMESPACE | grep kubearmor-prometheus-exporter | wc -l)
if [ $KUBEARMOR_EXPORTER != 0 ]; then
    kubectl delete -n $NAMESPACE -f exporter-deployment.yaml
fi
