package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"slack-app": func() (tfprotov6.ProviderServer, error) {
		return tfsdk.NewProtocol6Server(New("test")()), nil
	},
}

func testAccPreCheck(t *testing.T) {}
