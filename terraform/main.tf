terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "7tv"

    workspaces {
      prefix = "seventv-compactdisc-"
    }
  }
}

locals {
  infra_workspace_name = replace(terraform.workspace, "compactdisc", "infra")
  infra                = data.terraform_remote_state.infra.outputs
  image_url            = var.image_url != null ? var.image_url : format("ghcr.io/seventv/compactdisc:%s-latest", trimprefix(terraform.workspace, "seventv-compactdisc-"))
}
