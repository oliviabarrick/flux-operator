#!/bin/bash
set -e
echo "$SSH_KEY" > /tmp/ssh_key

# return 1 if the command returns 0, otherwise return 0
function not {
    ! $@
}

# wait for up to $1 seconds for some command to return true
function wait_for {
    set +x
    set +e

    max_tries=$1
    count=0
    ret=1


    while [ $count -lt $max_tries ] && [ $ret -ne 0 ]; do
        ${@:2}
        ret=$?
        sleep 1
        count=$(($count + 1))
    done

    set -e
    set -x

    return $ret
}

# clean up resources
function cleanup {
    set +e

    kubectl get all

    echo Clean up helm and testing resources before exiting
    kubectl delete -f deploy/
    kubectl delete secrets --all
    kubectl delete configmaps --all
    kubectl delete statefulsets --all
    kubectl delete deployments --all
    kubectl delete services --all
}

trap cleanup EXIT

set -x

MAXIMUM_TIMEOUT=600

echo Waiting for docker.
wait_for $MAXIMUM_TIMEOUT docker ps

echo Building Docker image.
docker build -t justinbarrick/flux-operator:latest .

echo Creating Flux resources.
kubectl create secret generic flux-git-example-deploy --from-file=identity=/tmp/ssh_key
kubectl apply -f deploy/flux-crd-namespaced.yaml
kubectl apply -f deploy/fluxhelmrelease-crd.yaml
kubectl apply -f deploy/k8s.yaml
kubectl apply -f deploy/cr.yaml
kubectl get deployments
kubectl get pods

echo Waiting for Flux to start up.

echo Confirming that all expected resources exist

wait_for $MAXIMUM_TIMEOUT kubectl get deployment flux-example
wait_for $MAXIMUM_TIMEOUT kubectl get deployment flux-example-memcached
wait_for $MAXIMUM_TIMEOUT kubectl get deployment flux-operator
wait_for $MAXIMUM_TIMEOUT kubectl get fluxhelmreleases memcached
kubectl logs $(kubectl get pods -l app=flux -o 'go-template={{ range .items }}{{ .metadata.name }}{{ "\n" }}{{ end }}' |head -n1)
wait_for $MAXIMUM_TIMEOUT kubectl get deployment nginx

echo Enabling the helm-operator
sed 's/enabled: false/enabled: true/g' deploy/cr.yaml |kubectl apply -f -

echo Waiting for helm-operator and tiller to start.

wait_for $MAXIMUM_TIMEOUT kubectl get deployment flux-example-helm-operator
wait_for $MAXIMUM_TIMEOUT kubectl get deployment tiller-deploy

echo Waiting for helm-operator to create resources.
wait_for $MAXIMUM_TIMEOUT kubectl get statefulset memcached-memcached

echo Disabling helm-operator
kubectl apply -f deploy/cr.yaml

echo Waiting for helm-operator to go away.
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example-helm-operator

kubectl delete -f deploy/cr.yaml
kubectl delete -f deploy/flux-crd-namespaced.yaml
kubectl delete -f deploy/fluxhelmrelease-crd.yaml
kubectl delete -f deploy/k8s.yaml
kubectl delete secret flux-git-example-deploy

echo Waiting for resources to clean up

wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-operator
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example-memcached
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment tiller-deploy

echo "Exiting with success!"
