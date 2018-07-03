#!/bin/sh

set -e

echo logging in..
docker login -u $DOCKER_USER -p $DOCKER_PASSWORD

set -x

docker tag justinbarrick/flux-operator:latest justinbarrick/flux-operator:$TAG
docker push justinbarrick/flux-operator:latest
docker push justinbarrick/flux-operator:$TAG
