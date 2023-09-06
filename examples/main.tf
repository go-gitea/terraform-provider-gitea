resource "gitea_repository" "test" {
  username     = "lerentis"
  name         = "test"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

resource "gitea_repository" "mirror" {
  username                     = "lerentis"
  name                         = "terraform-provider-gitea-mirror"
  description                  = "Mirror of Terraform Provider"
  mirror                       = true
  migration_clone_address      = "https://git.uploadfilter24.eu/lerentis/terraform-provider-gitea.git"
  migration_service            = "gitea"
  migration_service_auth_token = var.gitea_mirror_token
}

resource "gitea_org" "test_org" {
  name        = "test-org"
  description = "test description"
}

resource "gitea_repository" "org_repo" {
  username = gitea_org.test_org.name
  name     = "org-test-repo"
}

data "gitea_user" "me" {
  username = "lerentis"
}

resource "gitea_user" "test" {
  username             = "test"
  login_name           = "test"
  password             = "Geheim1!"
  email                = "test@user.dev"
  must_change_password = false
  admin                = true
}


resource "gitea_public_key" "test_user_key" {
  title     = "test"
  key       = file("${path.module}/resources/gitea_public_key/id_ed25519.pub")
  read_only = true
  username  = gitea_user.test.username
}


resource "gitea_team" "test_team" {
  name         = "Devs"
  organisation = gitea_org.test_org.name
  description  = "Devs of Test Org"
  permission   = "write"
  members      = [gitea_user.test.username]
}

resource "gitea_team" "admin_team" {
  name         = "Admins"
  organisation = gitea_org.test_org.name
  description  = "Admins of Test Org"
  permission   = "admin"
  members      = [data.gitea_user.me.username]
}

resource "gitea_git_hook" "org_repo_pre_receive" {
  name    = "pre-receive"
  user    = gitea_org.test_org.name
  repo    = gitea_repository.org_repo.name
  content = file("${path.module}/pre-receive.sh")
}

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

resource "gitea_token" "test_token" {
  username = data.gitea_user.me.username
  name     = "test-token"
}

resource "gitea_repository" "test_existing_user" {
  username     = "testuser2"
  name         = "testExistingUser"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

//resource "gitea_repository" "test_bs_user" {
//  username     = "manualTest"
//  name         = "testBullshitUser"
//  private      = true
//  issue_labels = "Default"
//  license      = "MIT"
//  gitignores   = "Go"
//}

output "token" {
  value = resource.gitea_token.test_token.token
  sensitive = true
}

data "gitea_repo" "org_repos" {
  name = each.key
  username = gitea_org.org1.name
  for_each = {
    for repo in resource.gitea_org.org1.repos : repo => repo
  }
}

output "repos" {
  value = data.gitea_repo.org_repos["repo1-in-org1"].clone_url
}
