provider "gitea" {
  base_url = var.gitea_url
  # Token Auth can not be used with this resource
  username = var.gitea_username
  password = var.gitea_password
}

resource "gitea_user" "test" {
  username             = "test"
  login_name           = "test"
  password             = "Geheim1!"
  email                = "test@user.dev"
  must_change_password = false
  admin                = true
}

resource "gitea_token" "test_token" {
  username = resource.gitea_user.test.username
  name     = "test-token"
}

output "token" {
  value     = resource.gitea_token.test_token.token
  sensitive = true
}
