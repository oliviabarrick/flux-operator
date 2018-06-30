package main

import (
	"context"
	"runtime"
	"os"

	stub "github.com/justinbarrick/flux-operator/pkg/stub"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	resource := "flux.codesink.net/v1alpha1"
	kind := "Flux"

	namespace := os.Getenv("FLUX_NAMESPACE")
	resyncPeriod := 5

	if namespace == "" {
		logrus.Infof("Watching for Fluxes at cluster scope.")
	} else {
		logrus.Infof("Watching for Fluxes in %s.", namespace)
	}

	sdk.Watch(resource, kind, "", resyncPeriod)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}
