resource "gitea_org" "test_org" {
  name = var.org_name
}

resource "gitea_repository" "org_repo" {
  username = gitea_org.test_org.name
  name     = var.repo_name
}

resource "gitea_user" "test_user" {
  password             = "Geheim1!"
  email                = "terraform@local.host"
  username             = var.user_name
  login_name           = var.user_name
  must_change_password = false
}

resource "gitea_repository" "user_repo" {
  username = gitea_user.test_user.username
  name     = var.repo_name
}

