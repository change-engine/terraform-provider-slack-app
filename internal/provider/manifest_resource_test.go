package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccManifestResource(t *testing.T) {
	isUrl, _ := regexp.Compile("https://.*")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManifestResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("slack-app_manifest.test", "oauth_authorize_url", isUrl),
				),
			},
			{
				ResourceName:            "slack-app_manifest.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials", "oauth_authorize_url"},
			},
			{
				Config: testAccManifestResourceConfig("two"),
			},
		},
	})
}

func testAccManifestResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`

resource "slack-app_manifest" "test" {
  manifest = jsonencode({
    display_information = {
      name = %[1]q
    }
	features = {
      bot_user = {
        display_name = "Test"
        always_online = false
      }
    }
    oauth_config = {
      redirect_urls = [
        "https://example.com/oauth"
      ]
      scopes = {
        bot = ["chat:write", "users:read.email", "users:read"]
      }
    }
    settings = {
      interactivity = {
        is_enabled = true
        request_url = "https://example.com/events"
      }
      org_deploy_enabled = false
      socket_mode_enabled = false
      token_rotation_enabled = false
    }
  })
}
`, configurableAttribute)
}
