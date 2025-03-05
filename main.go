package main

import (
	"context"

	"github.com/Cypherpunk-Labs/terraform-provider-mongodocs/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/cypherpunklabs/mongodocs",
	})
}
