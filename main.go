package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/provider"
)

// version is set by GoReleaser via ldflags on release builds.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "run the provider with debugger support (for example delve)")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/vetal-ca-org/voipms",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
