resource "gitea_repository" "repo" {
  username  = var.username
  name      = var.name
  auto_init = false
}

resource "gitea_repository_branch_protection" "main" {
  username = gitea_repository.repo.username
  name     = gitea_repository.repo.name

  rule_name             = "main"
  enable_push           = true
  status_check_patterns = var.branch_protection_patterns
}
