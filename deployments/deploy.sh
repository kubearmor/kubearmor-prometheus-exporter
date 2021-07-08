#!/bin/bash

NAMESPACE=kubearmor

if [ ! -z $1 ]; then
    NAMESPACE=$1
else
    echo "Default Namespace: $NAMESPACE"
fi

kubectl get namespace $NAMESPACE > /dev/null 2>&1
if [ $? != 0 ]; then
    kubectl create namespace $NAMESPACE
fi

KUBEARMOR_EXPORTER=$(kubectl get pods -n $NAMESPACE | grep kubearmor-prometheus-exporter | wc -l)
if [ $KUBEARMOR_EXPORTER == 0 ]; then
    kubectl apply -n $NAMESPACE -f exporter-deployment.yaml
fi
