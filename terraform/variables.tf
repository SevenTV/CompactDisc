variable "discord_docker_image" {
  type    = string
  default = "ghcr.io/seventv/compactdisc:latest"
}

data "terraform_remote_state" "infra" {
  backend = "remote"

  config = {
    organization = "7tv"
    workspaces = {
      name = "7tv-infra-${trimprefix(terraform.workspace, "7tv-discord-")}"
    }
  }
}

variable "website_url" {
  type = string
}

variable "cdn_url" {
  type = string
}

variable "discord_guild_id" {
  type = string
}

variable "discord_token" {
  type      = string
  sensitive = true
}
