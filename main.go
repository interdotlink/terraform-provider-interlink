package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"terraform-provider-interlink/internal/provider"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/interdotlink/interlink",
	}

	if err := providerserver.Serve(context.Background(), provider.New, opts); err != nil {
		log.Fatal(err.Error())
	}
}
