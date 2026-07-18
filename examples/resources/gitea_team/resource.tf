resource "gitea_org" "test_org" {
  name = "team-example-org"
}

resource "gitea_team" "test_team" {
  name         = "Devs"
  organisation = gitea_org.test_org.name
  description  = "Devs of Test Org"
  permission   = "write"
}

resource "gitea_repository" "test" {
  username     = gitea_org.test_org.name
  name         = "team-example-repo"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

resource "gitea_team" "test_team_restricted" {
  name                     = "Devs-Restricted"
  organisation             = gitea_org.test_org.name
  description              = "Restricted Devs of Test Org"
  permission               = "write"
  include_all_repositories = false
  repositories             = [gitea_repository.test.name]
}

resource "gitea_team" "test_team_units" {
  name         = "Devs-Granular-Permissions"
  organisation = gitea_org.test_org.name
  description  = "Devs with granular unit permissions"
  units_map = {
    "repo.code"   = "write"
    "repo.issues" = "read"
    "repo.wiki"   = "none"
  }
}
