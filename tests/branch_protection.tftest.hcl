# run "create_org" {
#   assert {
#     condition     = gitea_org.test_org.name == var.org_name
#     error_message = "${gitea_org.test_org.name} not eq ${var.org_name}"
#   }
# }
#
# run "create_repo" {
#   assert {
#     condition     = gitea_repository.org_repo.name == var.repo_name
#     error_message = "${gitea_repository.org_repo.name} not eq ${var.repo_name}"
#   }
# }

# run "create_user" {
#   assert {
#     condition     = gitea_user.test_user.username == var.user_name
#     error_message = "${gitea_user.test_user.username} not eq ${var.user_name}"
#   }
# }
#
# run "create_user_repo" {
#   assert {
#     condition     = gitea_repository.user_repo.name == var.repo_name
#     error_message = "${gitea_repository.user_repo.name} not eq ${var.repo_name}"
#   }
# }

run "setup" {
  module {
    source = "./setup"
  }
}

run "apply_defaults" {
  variables {
    rule_name = "apply_defaults"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.name == var.repo_name
    error_message = "${gitea_repository_branch_protection.bp.name} not eq ${var.repo_name}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.username == var.user_name
    error_message = "${gitea_repository_branch_protection.bp.username} not eq ${var.user_name}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.rule_name == var.rule_name
    error_message = "${gitea_repository_branch_protection.bp.rule_name} not eq ${var.rule_name}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.protected_file_patterns == var.protected_file_patterns
    error_message = "${gitea_repository_branch_protection.bp.protected_file_patterns} not eq ${var.protected_file_patterns}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.unprotected_file_patterns == var.unprotected_file_patterns
    error_message = "${gitea_repository_branch_protection.bp.unprotected_file_patterns} not eq ${var.unprotected_file_patterns}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_push == var.enable_push
    error_message = "${gitea_repository_branch_protection.bp.enable_push} not eq ${var.enable_push}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.push_whitelist_users == var.push_whitelist_users
    error_message = "${join(",", gitea_repository_branch_protection.bp.push_whitelist_users)} not eq ${join(",", var.push_whitelist_users)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.push_whitelist_teams == var.push_whitelist_teams
    error_message = "${join(",", gitea_repository_branch_protection.bp.push_whitelist_teams)} not eq ${join(",", var.push_whitelist_teams)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.push_whitelist_deploy_keys == var.push_whitelist_deploy_keys
    error_message = "${gitea_repository_branch_protection.bp.push_whitelist_deploy_keys} not eq ${var.push_whitelist_deploy_keys}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.require_signed_commits == var.require_signed_commits
    error_message = "${gitea_repository_branch_protection.bp.require_signed_commits} not eq ${var.require_signed_commits}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.required_approvals == var.required_approvals
    error_message = "${gitea_repository_branch_protection.bp.required_approvals} not eq ${var.required_approvals}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.approval_whitelist_users == var.approval_whitelist_users
    error_message = "${join(",", gitea_repository_branch_protection.bp.approval_whitelist_users)} not eq ${join(",", var.approval_whitelist_users)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.approval_whitelist_teams == var.approval_whitelist_teams
    error_message = "${join(",", gitea_repository_branch_protection.bp.approval_whitelist_teams)} not eq ${join(",", var.approval_whitelist_teams)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.dismiss_stale_approvals == var.dismiss_stale_approvals
    error_message = "${gitea_repository_branch_protection.bp.dismiss_stale_approvals} not eq ${var.dismiss_stale_approvals}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.status_check_patterns == var.status_check_patterns
    error_message = "${join(",", gitea_repository_branch_protection.bp.status_check_patterns)} not eq ${join(",", var.status_check_patterns)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.merge_whitelist_users == var.merge_whitelist_users
    error_message = "${join(",", gitea_repository_branch_protection.bp.merge_whitelist_users)} not eq ${join(",", var.merge_whitelist_users)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.merge_whitelist_teams == var.merge_whitelist_teams
    error_message = "${join(",", gitea_repository_branch_protection.bp.merge_whitelist_teams)} not eq ${join(",", var.merge_whitelist_teams)}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.block_merge_on_rejected_reviews == var.block_merge_on_rejected_reviews
    error_message = "${gitea_repository_branch_protection.bp.block_merge_on_rejected_reviews} not eq ${var.block_merge_on_rejected_reviews}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.block_merge_on_official_review_requests == var.block_merge_on_official_review_requests
    error_message = "${gitea_repository_branch_protection.bp.block_merge_on_official_review_requests} not eq ${var.block_merge_on_official_review_requests}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.block_merge_on_outdated_branch == var.block_merge_on_outdated_branch
    error_message = "${gitea_repository_branch_protection.bp.block_merge_on_outdated_branch} not eq ${var.block_merge_on_outdated_branch}"
  }
}

run "simple_params" {
  variables {
    rule_name                               = "simple_params"
    protected_file_patterns                 = "foobar.yaml"
    unprotected_file_patterns               = "foobar.yaml"
    require_signed_commits                  = true
    required_approvals                      = 10
    dismiss_stale_approvals                 = true
    block_merge_on_rejected_reviews         = true
    block_merge_on_official_review_requests = true
    block_merge_on_outdated_branch          = true
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.protected_file_patterns == var.protected_file_patterns
    error_message = "${gitea_repository_branch_protection.bp.protected_file_patterns} not eq ${var.protected_file_patterns}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.unprotected_file_patterns == var.unprotected_file_patterns
    error_message = "${gitea_repository_branch_protection.bp.unprotected_file_patterns} not eq ${var.unprotected_file_patterns}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.require_signed_commits == var.require_signed_commits
    error_message = "${gitea_repository_branch_protection.bp.require_signed_commits} not eq ${var.require_signed_commits}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.required_approvals == var.required_approvals
    error_message = "${gitea_repository_branch_protection.bp.required_approvals} not eq ${var.required_approvals}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.dismiss_stale_approvals == var.dismiss_stale_approvals
    error_message = "${gitea_repository_branch_protection.bp.dismiss_stale_approvals} not eq ${var.dismiss_stale_approvals}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.block_merge_on_rejected_reviews == var.block_merge_on_rejected_reviews
    error_message = "${gitea_repository_branch_protection.bp.block_merge_on_rejected_reviews} not eq ${var.block_merge_on_rejected_reviews}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.block_merge_on_official_review_requests == var.block_merge_on_official_review_requests
    error_message = "${gitea_repository_branch_protection.bp.block_merge_on_official_review_requests} not eq ${var.block_merge_on_official_review_requests}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.block_merge_on_outdated_branch == var.block_merge_on_outdated_branch
    error_message = "${gitea_repository_branch_protection.bp.block_merge_on_outdated_branch} not eq ${var.block_merge_on_outdated_branch}"
  }
}

run "enable_push" {
  variables {
    rule_name   = "enable_push"
    enable_push = true
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_push == var.enable_push
    error_message = "${gitea_repository_branch_protection.bp.enable_push} not eq ${var.enable_push}"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_push_whitelist == false
    error_message = "${gitea_repository_branch_protection.bp.enable_push_whitelist} not eq `false`"
  }
}

run "implicit_push_whitelist_with_users" {
  variables {
    enable_push          = true
    rule_name            = "implicit_and_push_whitelist_with_users"
    push_whitelist_users = ["test_user"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_push_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_push_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.push_whitelist_users == tolist(var.push_whitelist_users)
    error_message = "${join(",", gitea_repository_branch_protection.bp.push_whitelist_users)} not eq ${join(",", var.push_whitelist_users)}"
  }
}

run "implicit_push_whitelist_with_teams" {
  variables {
    rule_name            = "implicit_and_push_whitelist_with_teams"
    repo_user_name       = "test-org"
    enable_push          = true
    push_whitelist_teams = ["Owners"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_push_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_push_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.push_whitelist_teams == tolist(var.push_whitelist_teams)
    error_message = "${join(",", gitea_repository_branch_protection.bp.push_whitelist_teams)} not eq ${join(",", var.push_whitelist_teams)}"
  }
}

run "implicit_push_whitelist_with_deploy_keys" {
  variables {
    rule_name                  = "implicit_push_whitelist_with_deploy_keys"
    enable_push                = true
    push_whitelist_deploy_keys = true
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_push_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_push_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.push_whitelist_deploy_keys == var.push_whitelist_deploy_keys
    error_message = "${gitea_repository_branch_protection.bp.push_whitelist_deploy_keys} not eq ${var.push_whitelist_deploy_keys}"
  }
}

run "implicit_enable_approve_whitelist_with_users" {
  variables {
    rule_name                = "implicit_enable_approve_whitelist_with_users"
    approval_whitelist_users = ["test_user"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_approval_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_approval_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.approval_whitelist_users == tolist(var.approval_whitelist_users)
    error_message = "${join(",", gitea_repository_branch_protection.bp.approval_whitelist_users)} not eq ${join(",", var.approval_whitelist_users)}"
  }
}

run "implicit_enable_approve_whitelist_with_teams" {
  variables {
    rule_name                = "implicit_enable_approve_whitelist_with_teams"
    repo_user_name           = "test-org"
    approval_whitelist_teams = ["Owners"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_approval_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_approval_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.approval_whitelist_teams == tolist(var.approval_whitelist_teams)
    error_message = "${join(",", gitea_repository_branch_protection.bp.approval_whitelist_teams)} not eq ${join(",", var.approval_whitelist_teams)}"
  }
}

run "implicit_enable_merge_whitelist_with_users" {
  variables {
    rule_name             = "implicit_enable_merge_whitelist_with_users"
    merge_whitelist_users = ["test_user"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_merge_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_merge_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.merge_whitelist_users == tolist(var.merge_whitelist_users)
    error_message = "${join(",", gitea_repository_branch_protection.bp.merge_whitelist_users)} not eq ${join(",", var.merge_whitelist_users)}"
  }
}

run "implicit_enable_merge_whitelist_with_teams" {
  variables {
    rule_name             = "implicit_enable_merge_whitelist_with_teams"
    repo_user_name        = "test-org"
    merge_whitelist_teams = ["Owners"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_merge_whitelist == true
    error_message = "${gitea_repository_branch_protection.bp.enable_merge_whitelist} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.merge_whitelist_teams == tolist(var.merge_whitelist_teams)
    error_message = "${join(",", gitea_repository_branch_protection.bp.merge_whitelist_teams)} not eq ${join(",", var.merge_whitelist_teams)}"
  }
}

run "implicit_enable_status_check" {
  variables {
    rule_name             = "implicit_enable_status_check"
    status_check_patterns = ["terraform-tests", "tf fmt"]
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.enable_status_check == true
    error_message = "${gitea_repository_branch_protection.bp.enable_status_check} not eq `true`"
  }
  assert {
    condition     = gitea_repository_branch_protection.bp.status_check_patterns == tolist(var.status_check_patterns)
    error_message = "${join(",", gitea_repository_branch_protection.bp.status_check_patterns)} not eq ${join(",", var.status_check_patterns)}"
  }
}
