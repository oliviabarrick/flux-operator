// https://github.com/ant31/crd-validation

package main

import (
	"flag"
	"github.com/justinbarrick/flux-operator/pkg/installer"
	"github.com/justinbarrick/flux-operator/pkg/utils"
)

func main() {
	name := flag.String("name", "flux-operator", "Prefix to use for any resources created.")
	namespace := flag.String("namespace", "default", "Namespace to deploy flux-operator into.")
	watchNamespace := flag.String("watch-namespace", "", "If set, specifies the namespace to watch for Flux CRDs, if not set flux-operator watches all namespaces.")
	cluster := flag.Bool("cluster", false, "If set, creates a Cluster scoped CRD for Flux instead of Namespaced.")
	serviceAccount := flag.String("service-account", "", "Service account to use.")
	clusterRole := flag.String("cluster-role", "", "Cluster role to assign.")
	disableRbac := flag.Bool("disable-rbac", false, "Disable setting any RBAC settings.")
	gitSecret := flag.String("git-secret", "", "Default git secret name to use.")
	fluxOperatorImage := flag.String("flux-operator-image", utils.FluxOperatorImage, "Flux operator image name.")
	fluxOperatorVersion := flag.String("flux-operator-version", "latest", "Flux operator version name.")
	fluxImage := flag.String("flux-image", utils.FluxImage, "Flux image name.")
	fluxVersion := flag.String("flux-version", utils.FluxVersion, "Flux version name.")
	helmOperatorImage := flag.String("helm-operator-image", utils.HelmOperatorImage, "Helm-operator image name.")
	helmOperatorVersion := flag.String("helm-operator-version", utils.HelmOperatorVersion, "Helm-operator image version.")
	memcachedImage := flag.String("memcached-image", utils.MemcachedImage, "Memcached image name.")
	memcachedVersion := flag.String("memcached-version", utils.MemcachedVersion, "Memcached image version.")
	tillerImage := flag.String("tiller-image", utils.TillerImage, "Tiller image name.")
	tillerVersion := flag.String("tiller-version", utils.TillerVersion, "Tiller image version.")
	disableRoles := flag.Bool("disable-roles", false, "Do not allow flux-operator to assign roles.")
	disableClusterRoles := flag.Bool("disable-cluster-roles", false, "Do not allow flux-operator to assign cluster roles.")

	flag.Parse()

	installer.DryRun(installer.FluxOperatorConfig{
		Name:                *name,
		Namespace:           *namespace,
		Cluster:             *cluster,
		ServiceAccount:      *serviceAccount,
		ClusterRole:         *clusterRole,
		DisableRBAC:         *disableRbac,
		FluxNamespace:       *watchNamespace,
		GitSecret:           *gitSecret,
		FluxOperatorImage:   *fluxOperatorImage,
		FluxOperatorVersion: *fluxOperatorVersion,
		FluxImage:           *fluxImage,
		FluxVersion:         *fluxVersion,
		HelmOperatorImage:   *helmOperatorImage,
		HelmOperatorVersion: *helmOperatorVersion,
		TillerImage:         *tillerImage,
		TillerVersion:       *tillerVersion,
		MemcachedImage:      *memcachedImage,
		MemcachedVersion:    *memcachedVersion,
		DisableRoles:        *disableRoles,
		DisableClusterRoles: *disableClusterRoles,
	})
}
