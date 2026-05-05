# Enumerate every repo owned by an organization and emit their full names.
# Pair this with `for_each` to apply branch-protection (or any per-repo
# resource) across the whole org without hardcoding the repo list.

data "gitea_repos" "org_repos" {
  owner    = "my-org"
  archived = "false"
}

output "org_repo_full_names" {
  value = data.gitea_repos.org_repos.full_names
}

# Apply baseline branch protection to every non-archived repo in the org.
resource "gitea_repository_branch_protection" "baseline" {
  for_each = {
    for r in data.gitea_repos.org_repos.repos : r.full_name => r
  }

  username  = each.value.owner
  name      = each.value.name
  rule_name = each.value.default_branch
}
