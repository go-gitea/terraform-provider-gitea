terraform {
  required_providers {
    tls = {
      source = "hashicorp/tls"
      version = "4.0.4"
    }
  }
}

resource "tls_private_key" "example" {
  type = "RSA"
  rsa_bits = 4096
}

resource "gitea_repository" "example" {
  name = "example"
  private = true
}

resource "gitea_repository_key" "example" {
  repository = gitea_repository.example.id
  title = "Example Deploy Key"
  read_only = true
  key = tls_private_key.example.public_key_openssh
}