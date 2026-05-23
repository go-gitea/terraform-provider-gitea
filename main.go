package main // import "gitea.dev/terraform-provider-gitea"

import (
	"gitea.dev/terraform-provider-gitea/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

var Version = "development"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: gitea.Provider})
}
