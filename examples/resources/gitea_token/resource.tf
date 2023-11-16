provider "gitea" {
  base_url = var.gitea_url
  # Token Auth can not be used with this resource
  username = var.gitea_username
  password = var.gitea_password
}

// The token owner is the creator of the token
resource "gitea_token" "test_token" {
  name   = "test_token"
  scopes = ["all"]
}

output "token" {
  value     = resource.gitea_token.test_token.token
  sensitive = true
}
