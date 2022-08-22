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
