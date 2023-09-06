resource "gitea_org" "org1" {
  name = "org1"
}

resource "gitea_org" "org2" {
  name = "org2"
}

resource "gitea_repository" "repo1_in_org1" {
  username = gitea_org.org1.name
  name     = "repo1-in-org1"
}

resource "gitea_fork" "user_fork_of_repo1_in_org1" {
  owner = gitea_org.org1.name
  repo  = gitea_repository.repo1_in_org1.name
}

resource "gitea_fork" "org2_fork_of_repo1_in_org1" {
  owner        = gitea_org.org1.name
  repo         = gitea_repository.repo1_in_org1.name
  organization = gitea_org.org2.name
}
