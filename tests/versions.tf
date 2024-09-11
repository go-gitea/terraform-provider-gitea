terraform {
  required_providers {
    gitea = {
      source  = "go-gitea/gitea"
      version = "0.3.0"
    }
  }
  required_version = ">= 0.13"
}

provider "gitea" {
  base_url = "http://localhost:3000"
  username = "gitea_admin"
  password = "gitea_admin"
  insecure = true
}
