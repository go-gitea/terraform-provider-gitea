package main // import "code.gitea.ioterraform-provider-gitea"

import (
	"code.gitea.io/terraform-provider-gitea/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

var Version = "development"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: gitea.Provider})
}
