resource "gitea_repository_branch_protection" "bp" {
  username                                = var.repo_user_name
  name                                    = var.repo_name
  rule_name                               = var.rule_name
  protected_file_patterns                 = var.protected_file_patterns
  unprotected_file_patterns               = var.unprotected_file_patterns
  enable_push                             = var.enable_push
  push_whitelist_users                    = var.push_whitelist_users
  push_whitelist_teams                    = var.push_whitelist_teams
  push_whitelist_deploy_keys              = var.push_whitelist_deploy_keys
  require_signed_commits                  = var.require_signed_commits
  required_approvals                      = var.required_approvals
  approval_whitelist_users                = var.approval_whitelist_users
  approval_whitelist_teams                = var.approval_whitelist_teams
  dismiss_stale_approvals                 = var.dismiss_stale_approvals
  status_check_patterns                   = var.status_check_patterns
  merge_whitelist_users                   = var.merge_whitelist_users
  merge_whitelist_teams                   = var.merge_whitelist_teams
  block_merge_on_rejected_reviews         = var.block_merge_on_rejected_reviews
  block_merge_on_official_review_requests = var.block_merge_on_official_review_requests
  block_merge_on_outdated_branch          = var.block_merge_on_outdated_branch
}

# //
# // not implemented in go-gitea-sdk
# //
# // "ignore_stale_approvals": {
# // Description: `Do not count approvals that were made on older commits (stale reviews) towards how many approvals the PR has. Irrelevant if stale reviews are already dismissed.`,
# // },
