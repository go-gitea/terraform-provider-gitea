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

resource "gitea_repository_file" "example" {
  username       = gitea_user.example.username
  name           = gitea_repository.test.name
  file_path      = "README.md"
  content        = "# Example Repository\nThis is an example repository with a README updated via the `gitea_repository_file` resource."
  branch         = "main"
  commit_message = "Update README.md"
}

resource "gitea_repository_file" "example2" {
  username       = gitea_user.example.username
  name           = gitea_repository.test.name
  file_path      = "example.txt"
  content        = "# Example Text File\nThis is an example text file created via the `gitea_repository_file` resource."
  branch         = "main"
  commit_message = "Add example.txt"
}

resource "gitea_repository_file" "example_base64" {
  username       = gitea_user.example.username
  name           = gitea_repository.test.name
  file_path      = "base64.txt"
  encoding       = "base64"
  content        = base64encode("This string was sent to the resource provider as base64 encoded text")
  branch         = "main"
  commit_message = "Add base64.txt"
}
