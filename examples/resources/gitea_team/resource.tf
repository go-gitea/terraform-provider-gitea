resource "gitea_org" "test_org" {
  name = "test-org"
}

resource "gitea_user" "test" {
  username             = "test"
  login_name           = "test"
  password             = "Geheim1!"
  email                = "test@user.dev"
  must_change_password = false
  admin                = true
}


resource "gitea_team" "test_team" {
  name         = "Devs"
  organisation = gitea_org.test_org.name
  description  = "Devs of Test Org"
  permission   = "write"
  members      = [gitea_user.test.username]
}


resource "gitea_repository" "test" {
  username     = gitea_org.test_org.name
  name         = "test"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

resource "gitea_team" "test_team_restricted" {
  name                     = "Restricted Devs"
  organisation             = gitea_org.test_org.name
  description              = "Restricted Devs of Test Org"
  permission               = "write"
  members                  = [gitea_user.test.username]
  include_all_repositories = false
  repositories             = [gitea_repository.test.name]
}
