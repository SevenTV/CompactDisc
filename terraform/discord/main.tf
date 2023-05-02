resource "kubernetes_namespace" "discord" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_secret" "discord" {
  metadata {
    name      = "discord"
    namespace = kubernetes_namespace.discord.metadata[0].name
  }
  data = {
    "config.yaml" = templatefile("${path.module}/config.yaml", {
      website_url = var.website_url
      cdn_url = var.cdn_url
      discord_guild_id = var.discord_guild_id
      discord_token = var.discord_token
      mongo_db = "seventv"
      mongo_uri = "mongodb://root:${var.infra.mongo_password}@${var.infra.mongo_host}/?authSource=admin&readPreference=secondaryPreferred"
      redis_address = var.infra.redis_host
      redis_password = var.infra.redis_password
    })
  }
}

resource "kubernetes_deployment" "discord" {
  metadata {
    name = "discord"
    labels = {
      app = "discord"
    }
    namespace = kubernetes_namespace.discord.metadata[0].name
  }

  timeouts {
    create = "2m"
    update = "2m"
    delete = "2m"
  }

  spec {
    selector {
      match_labels = {
        app = "discord"
      }
    }
    template {
      metadata {
        labels = {
          app = "discord"
        }
      }
      spec {
        container {
          name              = "discord"
          image             = var.docker_image
          image_pull_policy = "Always"
          port {
            container_port = 3000
            name           = "api"
          }
          port {
            container_port = 9200
            name           = "health"
          }
          readiness_probe {
            initial_delay_seconds = 5
            period_seconds        = 5
            tcp_socket {
              port = "health"
            }
          }
          liveness_probe {
            initial_delay_seconds = 5
            period_seconds        = 5
            tcp_socket {
              port = "health"
            }
          }
          startup_probe {
            initial_delay_seconds = 5
            period_seconds        = 5
            tcp_socket {
              port = "health"
            }
          }
          security_context {
            allow_privilege_escalation = false
            privileged                 = false
            read_only_root_filesystem  = true
            run_as_non_root            = true
            run_as_user                = 1000
            run_as_group               = 1000
            capabilities {
              drop = ["ALL"]
            }
          }
          resources {
            limits = {
              "cpu"    = "250m"
              "memory" = "250Mi"
            }
            requests = {
              "cpu"    = "250m"
              "memory" = "250Mi"
            }
          }
          volume_mount {
            name       = "config"
            mount_path = "/app/config.yaml"
            sub_path   = "config.yaml"
            read_only  = true
          }
        }
        volume {
          name = "config"
          secret {
            secret_name  = kubernetes_secret.discord.metadata[0].name
            default_mode = "0644"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "discord" {
  metadata {
    name      = "discord"
    namespace = kubernetes_namespace.discord.metadata[0].name
  }

  spec {
    selector = {
      app = "discord"
    }
    port {
      name = "api"
      port = 3000
      target_port = "api"
    }
    port {
      name = "health"
      port = 9200
      target_port = "health"
    }
  }
}
