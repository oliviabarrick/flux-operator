// https://github.com/ant31/crd-validation

package main

import (
	"flag"
	crdutils "github.com/ant31/crd-validation/pkg"
	v1alpha1 "github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"os"
)

var (
	cfg crdutils.Config
)

func init() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset = crdutils.InitFlags(&cfg, flagset)
	flagset.Parse(os.Args[1:])
}

func main() {
	cfg.GetOpenAPIDefinitions = v1alpha1.GetOpenAPIDefinitions
	crd := crdutils.NewCustomResourceDefinition(cfg)
	crdutils.MarshallCrd(crd, cfg.OutputFormat)
	os.Exit(0)
}
