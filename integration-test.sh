#!/bin/bash
set -e
echo "$SSH_KEY" > /tmp/ssh_key

function cr {
    BRANCH="${DRONE_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}"
    sed "s#gitBranch: master#gitBranch: $BRANCH#g" $1
}

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
    cr deploy/cr-namespaced.yaml |kubectl delete -f -
    cr deploy/cr-cluster.yaml |kubectl delete -f -
    kubectl delete -f deploy/
    kubectl delete secrets --all
    kubectl delete configmaps --all
    kubectl delete statefulsets --all
    kubectl delete deployments --all
    kubectl delete services --all
    kubectl delete namespace lol
}

trap cleanup EXIT

set -x

MAXIMUM_TIMEOUT=600

echo Waiting for docker.
wait_for $MAXIMUM_TIMEOUT docker ps

echo Building Docker image.
docker build -t justinbarrick/flux-operator:latest .

echo Creating Flux resources.
kubectl create namespace lol
kubectl create secret generic flux-git-example-deploy --from-file=identity=/tmp/ssh_key
kubectl apply -f deploy/flux-crd-namespaced.yaml
kubectl apply -f deploy/fluxhelmrelease-crd.yaml
kubectl apply -f deploy/k8s.yaml
cr deploy/cr-namespaced.yaml | kubectl apply -f -
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
not kubectl get deployment -n lol nginx2

echo Enabling the helm-operator
cr deploy/cr-namespaced.yaml |sed 's/enabled: false/enabled: true/g' \
    |sed 's/flux-namespace: default/flux-namespace: ""/g' |kubectl apply -f -

echo Waiting for helm-operator and tiller to start.

wait_for $MAXIMUM_TIMEOUT kubectl get deployment flux-example-helm-operator
wait_for $MAXIMUM_TIMEOUT kubectl get deployment flux-example-tiller-deploy

echo Waiting for flux to create resources.
wait_for $MAXIMUM_TIMEOUT kubectl get deployment -n lol nginx2

echo Waiting for helm-operator to create resources.
wait_for $MAXIMUM_TIMEOUT kubectl get statefulset memcached-memcached

echo Disabling helm-operator
cr deploy/cr-namespaced.yaml |kubectl apply -f -

echo Waiting for helm-operator to go away.
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example-helm-operator

cr deploy/cr-namespaced.yaml |kubectl delete -f -
kubectl delete -f deploy/flux-crd-namespaced.yaml
kubectl delete -f deploy/fluxhelmrelease-crd.yaml
kubectl delete -f deploy/k8s.yaml
kubectl delete deployment nginx

echo Waiting for resources to clean up

wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-operator
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example-memcached
wait_for $MAXIMUM_TIMEOUT not kubectl get deployment flux-example-tiller-deploy

echo Starting cluster scoped flux

kubectl apply -f deploy/flux-crd-cluster.yaml
kubectl apply -f deploy/fluxhelmrelease-crd.yaml
kubectl apply -f deploy/k8s.yaml
cr deploy/cr-cluster.yaml | kubectl apply -f -

wait_for $MAXIMUM_TIMEOUT kubectl get deployment nginx
wait_for $MAXIMUM_TIMEOUT kubectl get deployment -n lol nginx2

kubectl delete secret flux-git-example-deploy

echo "######################################################"
echo "############## Exiting with success! #################"
echo "######################################################"
