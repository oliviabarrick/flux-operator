Operator for creating and managing instances of [Weaveworks flux](https://github.com/weaveworks/flux)

# Installation

To deploy to your cluster:

```
kubectl apply -f deploy/crd.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/operator.yaml
```

# Creating a Flux instance

To create a Flux instance, create a `Flux` CRD:

```
apiVersion: "flux.codesink.net/v1alpha1"
kind: "Flux"
metadata:
  name: example
spec:
  namespace: default
  gitUrl: git@github.com:justinbarrick/manifests
  args:
    connect: "ws://fluxcloud/"
```

This will create a flux pod called `flux-example` in the `default` namespace.

Settings:

* `namespace`: the namespace to deploy flux to.
* `gitUrl`: the URL to git repository to clone (required).
* `gitBranch`: the git branch to use (default: `master`).
* `gitPath`: the path with in the git repository to look for YAML in (default: `.`).
* `gitPollInterval`: the frequency with which to fetch the git repository (default: `5m0s`).
* `gitSecret`: the Kubernetes secret to use for cloning, if it does not exist it will
               be generated (default: `flux-$name-git-deploy`).
* `fluxImage`: the image to use for flux (default: `quay.io/weaveworks/flux`).
* `fluxVersion`: the version to use for flux (default: `1.4.0`).
* `clusterRole.enabled`: if enabled, a cluster role will be assigned to the service
                         account (default: `false`).
* `clusterRole.rules`: the list of rbac rules to use (default: full access to all resources).
* `role.enabled`: if enabled, a role will be assigned to the service
                  account (default: `false`).
* `role.rules`: the list of rbac rules to use (default: full access to all resources in the namespace).
* `args`: a map of args to pass to flux without `--` prepended.

You can also override some of the defaults by setting environment variables on the
operator itself:

* `FLUX_IMAGE`: the default image to use for flux.
* `FLUX_VERSION`: the default version to use for flux.
* `GIT_SECRET_NAME`: the git secret name to use.

# Git SSH key

If you already have an SSH key to use with flux, then create it:

```
kubectl create secret generic mysecret --from-file=identity=/home/user/.ssh/id_rsa
```

You can then reference the secret as `gitSecret: mysecret` in your Flux YAML.

To get the SSH public key from the SSH key you are using run:

```
ssh-keygen -y -f <(kubectl get secret -o 'go-template={{ .data.identity }}' flux-example-git-deploy |base64 -d)
```

You can then paste this into your Github deploy keys.

# RBAC

By default, no RBAC settings are created, but a service account is created and assigned
to the flux pod. The service account is called `flux-$fluxname`.

You can enable both a role and a cluster role on the service account.

To enable a cluster role, set `clusterRole.enabled`. The default cluster role created
will grant access to all resources in the cluster.

To enable a role, set `role.enabled`. The default role created will grant access to all
resources in the namespace.

You can also set custom RBAC rules by setting `role.rules` or `clusterRole.rules`:

```
apiVersion: "flux.codesink.net/v1alpha1"
kind: "Flux"
metadata:
  name: example
  namespace: default
spec:
  gitUrl: git@github.com:justinbarrick/manifests
  clusterRole:
    enabled: true
    rules:
      - apiGroups: ['*']
        resources: ['*']
        verbs: ['*']
      - nonResourceURLs: ['*']
        verbs: ['*']
```

See [Kubernetes role documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole) for more information.

# TODO

Next steps:

* Add support for Helm and the Helm operator.
