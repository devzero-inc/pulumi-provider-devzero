package main

import (
	"context"
	"fmt"
	"os"

	devprovider "github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/provider"
	pulumiprovider "github.com/pulumi/pulumi-go-provider"
)

// version is injected at build time via -ldflags "-X main.version=<semver>".
var version = "dev"

func main() {
	if err := pulumiprovider.RunProvider(context.Background(), "devzero", version, devprovider.New()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
