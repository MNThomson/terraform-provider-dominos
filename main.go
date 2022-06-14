package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/mnthomson/terraform-provider-dominos/internal/provider"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	version string = "dev"
)

func main() {
	opts := &plugin.ServeOpts{ProviderFunc: provider.Provider}

	plugin.Serve(opts)
}
