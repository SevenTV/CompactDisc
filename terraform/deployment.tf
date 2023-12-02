data "kubernetes_namespace" "app" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_secret" "app" {
  metadata {
    name      = "api"
    namespace = var.namespace
  }

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      discord_guild_id        = var.discord_guild_id
      discord_default_role_id = var.discord_default_role_id
      discord_bot_token       = var.discord_bot_token
      ch_activity_feed        = "817375925271527449"
      ch_mod_logs             = "989251544165777450"
      ch_mod_actor_tracker    = "1080982942156869743"
      ch_events               = "1015281319758004335"
      mongo_uri               = local.infra.mongodb_uri
      mongo_username          = local.infra.mongodb_user_app.username
      mongo_password          = local.infra.mongodb_user_app.password
      mongo_database          = "7tv"
      redis_username          = "default"
      redis_password          = local.infra.redis_password
    })
  }
}
