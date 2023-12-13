package main

import (
    "context"
    "log"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    provider "terraform-provider-interlink/provider"
)

func main() {
    opts := providerserver.ServeOpts{
        Address: "inter.link/tech/interlink",
    }

    err := providerserver.Serve(context.Background(), provider.New(), opts)
    if err != nil {
        log.Fatal(err.Error())
    }
}
