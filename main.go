package main

import (
	"context"

	"github.com/change-engine/terraform-provider-slack-app/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate tofu fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name slack-app

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/change-engine/slack-app",
	})
}
