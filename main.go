package main

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0 generate --provider-name artifactapi

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/IUCLabs/terraform-provider-artifactapi/internal/provider"
)

var version = "0.0.1"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/iuclabs/artifactapi",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
