# https://github.com/ant31/crd-validation
OPENAPI_GEN := $(shell command -v openapi-gen 2> /dev/null)
REPO := github.com/justinbarrick/flux-operator
ifeq ($(GOBIN),)
GOBIN :=${GOPATH}/bin
endif
DATE := $(shell date '+%Y-%m-%d %H:%M:%S')

all: generate-crds

$(GOBIN)/openapi-gen:
	go get -u -v -d k8s.io/code-generator/cmd/openapi-gen
	cd $(GOPATH)/src/k8s.io/code-generator; git checkout release-1.8
	go install k8s.io/code-generator/cmd/openapi-gen

pkg/apis/flux/v1alpha1/openapi_generated.go: $(GOBIN)/openapi-gen
	openapi-gen -i $(REPO)/pkg/apis/flux/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1 -p $(REPO)/pkg/apis/flux/v1alpha1 --go-header-file="$(GOPATH)/src/github.com/justinbarrick/flux-operator/.header"

generate-openapi: pkg/apis/flux/v1alpha1/openapi_generated.go

deploy/flux-operator-namespaced.yaml:
	./fluxopctl > deploy/flux-operator-namespaced.yaml

deploy/flux-operator-cluster.yaml:
	./fluxopctl -cluster > deploy/flux-operator-cluster.yaml

fluxopctl:
	CGO_ENABLED=0 go build -ldflags '-w -s' -installsuffix cgo -o fluxopctl cmd/fluxopctl/main.go

generate-crds: clean generate-openapi fluxopctl deploy/flux-operator-namespaced.yaml deploy/flux-operator-cluster.yaml

.PHONY: openapi-gen build test clean install all

install: $(GOBIN)/flux-operator-crd-gen

test:
	test -z $(shell gofmt -l ./cmd ./pkg)
	go test github.com/justinbarrick/flux-operator/...

build:
	gofmt -w ./cmd ./pkg
	CGO_ENABLED=0 go build -ldflags '-w -s' -installsuffix cgo -o flux-operator cmd/flux-operator/main.go

clean:
	rm -f pkg/apis/flux/v1alpha1/openapi_generated.go
	rm -f deploy/flux-operator-namespaced.yaml
	rm -f deploy/flux-operator-cluster.yaml
	rm -f ./flux-operator
