Operator for creating and managing instances of [Weaveworks flux](https://github.com/weaveworks/flux) and [tiller](https://github.com/kubernetes/helm).

Use-cases:

* Doing GitOps without a monorepo. You can easily split your manifests into repos per team, project or namespace.
* Simplify the deployment of your Flux and Tiller (and, in the future, helm-operator) instances.
* Easily manage Flux RBAC policies to prevent any single Flux instance from having access to the entire cluster.
* Use Helm without violating your RBAC policies.

# Installation

To deploy to your cluster:

```
kubectl apply -f deploy/k8s.yaml
```

# Creating a Flux instance

To create a Flux instance, create a `Flux` CR:

```
apiVersion: flux.codesink.net/v1alpha1
kind: Flux
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
               be generated (default: `flux-$name-git-deploy` or `$GIT_SECRET_NAME`).
* `fluxImage`: the image to use for flux (default: `quay.io/weaveworks/flux` or `$FLUX_IMAGE`).
* `fluxVersion`: the version to use for flux (default: `1.4.0` or `$FLUX_VERSION`).
* `clusterRole.enabled`: if enabled, a cluster role will be assigned to the service
                         account (default: `false`).
* `clusterRole.rules`: the list of rbac rules to use (default: full access to all resources).
* `role.enabled`: if enabled, a role will be assigned to the service
                  account (default: `false`).
* `role.rules`: the list of rbac rules to use (default: full access to all resources in the namespace).
* `tiller.enabled`: whether or not to deploy a tiller instance in the same namespace (default: false).
* `tiller.tillerImage`: the image to use with tiller (default: `gcr.io/kubernetes-helm/tiller` or `$TILLER_IMAGE`)
* `tiller.tillerVersion`: the image version to use with tiller (default: `v2.9.1` or `$TILLER_VERSION`)
* `args`: a map of args to pass to flux without `--` prepended.

You can also override some of the defaults by setting environment variables on the
operator itself:

* `FLUX_IMAGE`: the default image to use for flux.
* `FLUX_VERSION`: the default version to use for flux.
* `GIT_SECRET_NAME`: the git secret name to use.

# Git SSH key

If you already have an SSH key to use with flux, then add it as a secret to Kubernetes:

```
kubectl create secret generic mysecret --from-file=identity=/home/user/.ssh/id_rsa
```

You can then reference the secret as `gitSecret: mysecret` in your Flux YAML.

To get the SSH public key from the SSH key you are using run:

```
ssh-keygen -y -f <(kubectl get secret -o 'go-template={{ .data.identity }}' flux-example-git-deploy |base64 -d)
```

You can then paste this into your Github deploy keys.

# Creating a Tiller instance

The flux-operator can also deploy Tiller with Flux into your namespace. The Tiller
instance uses the same service account as Flux.

To enable, set `tiller.enabled` to true:

```
apiVersion: flux.codesink.net/v1alpha1
kind: Flux
metadata:
  name: example
spec:
  namespace: default
  gitUrl: git@github.com:justinbarrick/manifests
  role:
    enabled: true
  tiller:
    enabled: true
```

You can now securely use Helm in this namespace:

```
helm --tiller-namespace default ls
helm --tiller-namespace default install stable/memcached
```

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

* Add support for the Helm operator.
