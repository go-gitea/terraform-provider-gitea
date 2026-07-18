resource "gitea_user" "example" {
  username   = "repo_example_user"
  login_name = "repo_example_user"
  password   = "Geheim1!"
  email      = "repo_example_user@user.dev"
}

resource "gitea_repository" "test-repository" {
  name     = "test-repository"
  username = gitea_user.example.username
}

resource "gitea_repository_branch" "test-branch" {
  name       = "feat/testing-branch"
  repository = gitea_repository.test-repo.id
}

