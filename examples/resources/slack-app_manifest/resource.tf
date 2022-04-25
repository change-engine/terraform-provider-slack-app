resource "slack-app_manifest" "example" {
  manifest = jsonencode({
    display_information = {
      name = "Example App"
    }
    features = {
      bot_user = {
        display_name  = "Example"
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
        is_enabled  = true
        request_url = "https://example.com/events"
      }
      org_deploy_enabled     = false
      socket_mode_enabled    = false
      token_rotation_enabled = false
    }
  })
}
