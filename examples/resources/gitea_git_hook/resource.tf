resource "gitea_org" "test_org" {
  name = "test-org"
}

resource "gitea_repository" "org_repo" {
  username = gitea_org.test_org.name
  name     = "org-test-repo"
}

resource "gitea_git_hook" "org_repo_post_receive" {
  name    = "post-receive"
  user    = gitea_org.test_org.name
  repo    = gitea_repository.org_repo.name
  content = file("${path.module}/post-receive.sh")
}