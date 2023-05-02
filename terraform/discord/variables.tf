variable "docker_image" {
  type = string
}

variable "namespace" {
  type    = string
  default = "compactdisc"
}

variable "infra" {
  type      = any
  sensitive = true
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
