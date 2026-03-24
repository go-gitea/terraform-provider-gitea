resource "gitea_user" "repo_owner" {
  username   = "collab_example_owner"
  login_name = "collab_example_owner"
  password   = "Geheim1!"
  email      = "collab_example_owner@user.dev"
}

resource "gitea_user" "collaborator" {
  username   = "collab_example_user"
  login_name = "collab_example_user"
  password   = "Geheim1!"
  email      = "collab_example_user@user.dev"
}

resource "gitea_repository" "example" {
  username = gitea_user.repo_owner.username
  name     = "collab-example-repo"
  private  = true
}

resource "gitea_repository_collaborator" "example" {
  owner      = gitea_user.repo_owner.username
  repo       = gitea_repository.example.name
  username   = gitea_user.collaborator.username
  permission = "write"
}
