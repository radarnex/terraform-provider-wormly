package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	wormlyProvider "github.com/radarnex/terraform-provider-wormly/internal/provider"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/radarnex/wormly",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), func() provider.Provider {
		return wormlyProvider.New(version)
	}, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
