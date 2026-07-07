package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"terraform-provider-interlink/internal/provider"
)

// Run "go generate ./..." to regenerate the registry docs from the schema
// descriptions and the examples/ directory into docs/.
//go:generate go tool tfplugindocs generate --provider-name interlink

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/interdotlink/interlink",
	}

	if err := providerserver.Serve(context.Background(), provider.New, opts); err != nil {
		log.Fatal(err.Error())
	}
}
