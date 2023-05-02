terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "7tv"

    workspaces {
      prefix = "7tv-discord-"
    }
  }
}

module "discord" {
  source           = "./discord"
  cdn_url          = var.cdn_url
  docker_image     = var.discord_docker_image
  discord_guild_id = var.discord_guild_id
  discord_token    = var.discord_token
  website_url      = var.website_url
  infra            = data.terraform_remote_state.infra.outputs
}
