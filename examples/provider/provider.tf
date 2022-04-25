terraform {
  required_providers {
    slack-token = {
      source  = "change-engine/slack-token"
      version = "~> 0.1"
    }
    slack-app = {
      source  = "change-engine/slack-app"
      version = "~> 0.1"
    }
  }
}

resource "slack-token_refresh" "example" {}

provider "slack-app" {
  token = slack-token_refresh.example.token
}
