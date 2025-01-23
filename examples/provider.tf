terraform {
  required_providers {
    gitea = {
      source = "go-gitea/gitea"
      version = "0.6.0"
    }
  }
}

provider "gitea" {
  base_url = var.gitea_url
  username = "lerentis"
  password = var.gitea_password
  #token    = var.gitea_token
}
