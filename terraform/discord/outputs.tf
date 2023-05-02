output "hostname" {
  value = "${kubernetes_service.discord.metadata[0].name}.${var.namespace}.svc.cluster.local"
}
