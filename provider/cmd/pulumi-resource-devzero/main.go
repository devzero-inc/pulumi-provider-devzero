package main

import (
	"context"
	"fmt"
	"os"

	devprovider "github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/provider"
	pulumiprovider "github.com/pulumi/pulumi-go-provider"
)

func main() {
	if err := pulumiprovider.RunProvider(context.Background(), "devzero", "0.1.0", devprovider.New()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
