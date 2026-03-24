resource "gitea_user" "example" {
  username   = "repo_example_user"
  login_name = "repo_example_user"
  password   = "Geheim1!"
  email      = "repo_example_user@user.dev"
}

resource "gitea_repository" "test" {
  username     = gitea_user.example.username
  name         = "repository-test"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

data "gitea_repository_file" "readme" {
  username  = gitea_user.example.username
  name      = gitea_repository.test.name
  file_path = "README.md"
  branch    = "main"
}

output "readme" {
  value = base64decode(data.gitea_repository_file.readme.content)
}
