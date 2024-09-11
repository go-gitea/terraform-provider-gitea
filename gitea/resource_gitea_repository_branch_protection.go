package gitea

import (
	"log"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	repoBPUsername string = "username"
	repoBPName     string = "name"
	repoBPRuleName string = "rule_name"

	repoBPProtectedFilePatterns   string = "protected_file_patterns"
	repoBPUnprotectedFilePatterns string = "unprotected_file_patterns"

	repoBPEnablePush              string = "enable_push"
	repoBPEnablePushWhitelist     string = "enable_push_whitelist"
	repoBPPushWhitelistUsers      string = "push_whitelist_users"
	repoBPPushWhitelistTeams      string = "push_whitelist_teams"
	repoBPPushWhitelistDeployKeys string = "push_whitelist_deploy_keys"

	repoBPRequireSignedCommits string = "require_signed_commits"

	repoBPRequiredApprovals       string = "required_approvals"
	repoBPEnableApprovalWhitelist string = "enable_approval_whitelist"
	repoBPApprovalWhitelistUsers  string = "approval_whitelist_users"
	repoBPApprovalWhitelistTeams  string = "approval_whitelist_teams"
	repoBPDismissStaleApprovals   string = "dismiss_stale_approvals"
	// not implemented in go-gitea-sdk
	// repoBPIgnoreStaleApprovals   string = "ignore_stale_approvals"

	repoBPEnableStatusCheck   string = "enable_status_check"
	repoBPStatusCheckPatterns string = "status_check_patterns"

	repoBPEnableMergeWhitelist string = "enable_merge_whitelist"
	repoBPMergeWhitelistUsers  string = "merge_whitelist_users"
	repoBPMergeWhitelistTeams  string = "merge_whitelist_teams"

	repoBPBlockMergeOnRejectedReviews        string = "block_merge_on_rejected_reviews"
	repoBPBlockMergeOnOfficialReviewRequests string = "block_merge_on_official_review_requests"
	repoBPBlockMergeOnOutdatedBranch         string = "block_merge_on_outdated_branch"

	repoBPUpdatedAt string = "updated_at"
	repoBPCreatedAt string = "created_at"
)

func resourceRepositoryBranchProtectionRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	rule_name := d.Get(repoBPRuleName).(string)

	bp, resp, err := client.GetBranchProtection(user, repo, rule_name)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return
		} else {
			return err
		}
	}

	err = setRepositoryBranchProtectionData(bp, user, repo, d)
	return err
}

func generateWhitelist(d *schema.ResourceData, listname string) (enabled bool, users []string, teams []string) {
	u := d.Get(listname + "_users")
	users = make([]string, 0)
	if u != nil {
		for _, element := range u.([]interface{}) {
			users = append(users, element.(string))
		}
	}

	t := d.Get(listname + "_teams")
	teams = make([]string, 0)
	if u != nil {
		for _, element := range t.([]interface{}) {
			teams = append(teams, element.(string))
		}
	}

	if c := len(users) + len(teams); c > 0 {
		enabled = true
	}
	if listname == "push_whitelist" && d.Get(repoBPPushWhitelistDeployKeys).(bool) {
		enabled = true
	}

	log.Println("enabled?:", enabled, listname)
	return enabled, users, teams
}

func resourceRepositoryBranchProtectionCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)

	enablePushWhitelist, pushWhitelistUsernames, pushWhitelistTeams := generateWhitelist(d, "push_whitelist")
	enableMergeWhitelist, mergeWhitelistUsernames, mergeWhitelistTeams := generateWhitelist(d, "merge_whitelist")
	enableApprovalsWhitelist, approvalsWhitelistUsernames, approvalsWhitelistTeams := generateWhitelist(d, "approval_whitelist")

	statusCheckContexts := make([]string, 0)
	for _, element := range d.Get(repoBPStatusCheckPatterns).([]interface{}) {
		statusCheckContexts = append(statusCheckContexts, element.(string))
	}

	log.Println("create_ulist:", pushWhitelistUsernames)

	enableStatusCheck := false
	if len(statusCheckContexts) > 0 {
		enableStatusCheck = true
	}

	bpOption := gitea.CreateBranchProtectionOption{
		// BranchName is deprecated in gitea, but still required in go-gitea-sdk, therefore using RuleName
		BranchName:                    d.Get(repoBPRuleName).(string),
		RuleName:                      d.Get(repoBPRuleName).(string),
		EnablePush:                    d.Get(repoBPEnablePush).(bool),
		EnablePushWhitelist:           enablePushWhitelist,
		PushWhitelistUsernames:        pushWhitelistUsernames,
		PushWhitelistTeams:            pushWhitelistTeams,
		PushWhitelistDeployKeys:       d.Get(repoBPPushWhitelistDeployKeys).(bool),
		EnableMergeWhitelist:          enableMergeWhitelist,
		MergeWhitelistUsernames:       mergeWhitelistUsernames,
		MergeWhitelistTeams:           mergeWhitelistTeams,
		EnableStatusCheck:             enableStatusCheck,
		StatusCheckContexts:           statusCheckContexts,
		RequiredApprovals:             int64(d.Get(repoBPRequiredApprovals).(int)),
		EnableApprovalsWhitelist:      enableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   approvalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       approvalsWhitelistTeams,
		BlockOnRejectedReviews:        d.Get(repoBPBlockMergeOnRejectedReviews).(bool),
		BlockOnOfficialReviewRequests: d.Get(repoBPBlockMergeOnOfficialReviewRequests).(bool),
		BlockOnOutdatedBranch:         d.Get(repoBPBlockMergeOnOutdatedBranch).(bool),
		DismissStaleApprovals:         d.Get(repoBPDismissStaleApprovals).(bool),
		// IgnoreStaleApprovals:          d.Get(repoBPIgnoreStaleApprovals).(bool),
		RequireSignedCommits:    d.Get(repoBPRequireSignedCommits).(bool),
		ProtectedFilePatterns:   d.Get(repoBPProtectedFilePatterns).(string),
		UnprotectedFilePatterns: d.Get(repoBPUnprotectedFilePatterns).(string),
	}

	bp, _, err := client.CreateBranchProtection(user, repo, bpOption)
	if err != nil {
		return err
	}

	err = setRepositoryBranchProtectionData(bp, user, repo, d)
	return err
}

func resourceRepositoryBranchProtectionUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	rule_name := d.Id()

	enablePushWhitelist, pushWhitelistUsernames, pushWhitelistTeams := generateWhitelist(d, "push_whitelist")
	enableMergeWhitelist, mergeWhitelistUsernames, mergeWhitelistTeams := generateWhitelist(d, "merge_whitelist")
	enableApprovalsWhitelist, approvalsWhitelistUsernames, approvalsWhitelistTeams := generateWhitelist(d, "approval_whitelist")

	statusCheckContexts := make([]string, 0)
	for _, element := range d.Get(repoBPStatusCheckPatterns).([]interface{}) {
		statusCheckContexts = append(statusCheckContexts, element.(string))
	}

	enablePush := false
	if enablePushWhitelist == true || d.Get(repoBPEnablePush).(bool) == true {
		enablePush = true
	}
	pushWhitelistDeployKeys := d.Get(repoBPPushWhitelistDeployKeys).(bool)
	enableStatusCheck := false
	if len(statusCheckContexts) > 0 {
		enableStatusCheck = true
	}
	requiredApprovals := int64(d.Get(repoBPRequiredApprovals).(int))
	blockOnRejectedReviews := d.Get(repoBPBlockMergeOnRejectedReviews).(bool)
	blockOnOfficialReviewRequests := d.Get(repoBPBlockMergeOnOfficialReviewRequests).(bool)
	blockOnOutdatedBranch := d.Get(repoBPBlockMergeOnOutdatedBranch).(bool)
	dismissStaleApprovals := d.Get(repoBPDismissStaleApprovals).(bool)
	// ignoreStaleApprovals := d.Get(repoBPIgnoreStaleApprovals).(bool)
	requireSignedCommits := d.Get(repoBPRequireSignedCommits).(bool)
	protectedFilePatterns := d.Get(repoBPProtectedFilePatterns).(string)
	unprotectedFilePatterns := d.Get(repoBPUnprotectedFilePatterns).(string)

	bpOption := gitea.EditBranchProtectionOption{
		EnablePush:                    &enablePush,
		EnablePushWhitelist:           &enablePushWhitelist,
		PushWhitelistUsernames:        pushWhitelistUsernames,
		PushWhitelistTeams:            pushWhitelistTeams,
		PushWhitelistDeployKeys:       &pushWhitelistDeployKeys,
		EnableMergeWhitelist:          &enableMergeWhitelist,
		MergeWhitelistUsernames:       mergeWhitelistUsernames,
		MergeWhitelistTeams:           mergeWhitelistTeams,
		EnableStatusCheck:             &enableStatusCheck,
		StatusCheckContexts:           statusCheckContexts,
		RequiredApprovals:             &requiredApprovals,
		EnableApprovalsWhitelist:      &enableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   approvalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       approvalsWhitelistTeams,
		BlockOnRejectedReviews:        &blockOnRejectedReviews,
		BlockOnOfficialReviewRequests: &blockOnOfficialReviewRequests,
		BlockOnOutdatedBranch:         &blockOnOutdatedBranch,
		DismissStaleApprovals:         &dismissStaleApprovals,
		// IgnoreStaleApprovals:          &ignoreStaleApprovals,
		RequireSignedCommits:    &requireSignedCommits,
		ProtectedFilePatterns:   &protectedFilePatterns,
		UnprotectedFilePatterns: &unprotectedFilePatterns,
	}

	bp, _, err := client.EditBranchProtection(user, repo, rule_name, bpOption)
	if err != nil {
		return err
	}

	err = setRepositoryBranchProtectionData(bp, user, repo, d)
	return err
}

func resourceRepositoryBranchProtectionDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	rule_name := d.Id()

	_, err = client.DeleteBranchProtection(user, repo, rule_name)
	if err != nil {
		return err
	}
	return err
}

func setRepositoryBranchProtectionData(bp *gitea.BranchProtection, user string, repo string, d *schema.ResourceData) (err error) {
	d.SetId(bp.RuleName)
	d.Set(repoBPUsername, user)
	d.Set(repoBPName, repo)
	d.Set(repoBPProtectedFilePatterns, bp.ProtectedFilePatterns)
	d.Set(repoBPUnprotectedFilePatterns, bp.UnprotectedFilePatterns)
	d.Set(repoBPEnablePush, bp.EnablePush)
	d.Set(repoBPEnablePushWhitelist, bp.EnablePushWhitelist)
	d.Set(repoBPPushWhitelistUsers, bp.PushWhitelistUsernames)
	d.Set(repoBPPushWhitelistTeams, bp.PushWhitelistTeams)
	d.Set(repoBPPushWhitelistDeployKeys, bp.PushWhitelistDeployKeys)
	d.Set(repoBPRequireSignedCommits, bp.RequireSignedCommits)
	d.Set(repoBPRequiredApprovals, bp.RequiredApprovals)
	d.Set(repoBPEnableApprovalWhitelist, bp.EnableApprovalsWhitelist)
	d.Set(repoBPApprovalWhitelistUsers, bp.ApprovalsWhitelistUsernames)
	d.Set(repoBPApprovalWhitelistTeams, bp.ApprovalsWhitelistTeams)
	d.Set(repoBPDismissStaleApprovals, bp.DismissStaleApprovals)
	// d.Set(repoBPIgnoreStaleApprovals, bp.IgnoreStaleApprovals)
	d.Set(repoBPEnableStatusCheck, bp.EnableStatusCheck)
	d.Set(repoBPStatusCheckPatterns, bp.StatusCheckContexts)
	d.Set(repoBPEnableMergeWhitelist, bp.EnableMergeWhitelist)
	d.Set(repoBPMergeWhitelistUsers, bp.MergeWhitelistUsernames)
	d.Set(repoBPMergeWhitelistTeams, bp.MergeWhitelistTeams)
	d.Set(repoBPBlockMergeOnRejectedReviews, bp.BlockOnRejectedReviews)
	d.Set(repoBPBlockMergeOnOfficialReviewRequests, bp.BlockOnOfficialReviewRequests)
	d.Set(repoBPBlockMergeOnOutdatedBranch, bp.BlockOnOutdatedBranch)
	d.Set(repoBPUpdatedAt, bp.Updated)
	d.Set(repoBPCreatedAt, bp.Created)

	return err
}

func resourceGiteaRepositoryBranchProtection() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepositoryBranchProtectionRead,
		Create: resourceRepositoryBranchProtectionCreate,
		Update: resourceRepositoryBranchProtectionUpdate,
		Delete: resourceRepositoryBranchProtectionDelete,
		// TODO: importer ?
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User name or organization name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name",
			},
			"rule_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Protected Branch Name Pattern",
			},
			"protected_file_patterns": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Default:     "",
				Description: "Protected file patterns (separated using semicolon ';')",
			},
			"unprotected_file_patterns": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Default:     "",
				Description: "Unprotected file patterns (separated using semicolon ';')",
			},
			"enable_push": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Anyone with write access will be allowed to push to this branch
								(but not force push), add a whitelist users or teams to limit
								access.`,
			},
			"enable_push_whitelist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a push whitelist is used.",
			},
			"push_whitelist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				RequiredWith: []string{"enable_push"},
				Optional:     true,
				ForceNew:     false,
				Description:  "Allowlisted users for pushing. Requires enable_push to be set to true.",
			},
			"push_whitelist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				RequiredWith: []string{"enable_push"},
				Optional:     true,
				ForceNew:     false,
				Description:  "Allowlisted teams for pushing. Requires enable_push to be set to true.",
			},
			"push_whitelist_deploy_keys": {
				Type:         schema.TypeBool,
				RequiredWith: []string{"enable_push"},
				Optional:     true,
				ForceNew:     false,
				Default:      false,
				Description:  "Allow deploy keys with write access to push. Requires enable_push to be set to true.",
			},
			"require_signed_commits": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Reject pushes to this branch if they are unsigned or unverifiable.",
			},
			"required_approvals": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Default:     0,
				Description: "Allow only to merge pull request with enough positive reviews.",
			},
			"enable_approval_whitelist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a approval whitelist is used.",
			},
			"approval_whitelist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
				Description: `Only reviews from allowlisted users will count to the required
								approvals. Without approval allowlist, reviews from anyone with
								write access count to the required approvals.`,
			},
			"approval_whitelist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
				Description: `Only reviews from allowlisted teams will count to the required
								approvals. Without approval allowlist, reviews from anyone with
								write access count to the required approvals.`,
			},
			"dismiss_stale_approvals": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `When new commits that change the content of the pull request
								are pushed to the branch, old approvals will be dismissed.`,
			},
			//
			// not implemented in go-gitea-sdk
			//
			// "ignore_stale_approvals": {
			// 	Type:        schema.TypeBool,
			//  Optional 	 true,
			// 	ForceNew:    false,
			// 	Default:     false,
			// 	Description: `Do not count approvals that were made on older commits (stale
			//					reviews) towards how many approvals the PR has. Irrelevant if
			//					stale reviews are already dismissed.`,
			// },
			"enable_status_check": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: `Require status checks to pass before merging. When enabled,
								commits must first be pushed to another branch, then merged
								or pushed directly to a branch that matches this rule after
								status checks have passed. If no contexts are matched, the
								last commit must be successful regardless of context`,
			},
			"status_check_patterns": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
				Description: `Enter patterns to specify which status checks must pass before
								branches can be merged into a branch that matches this rule.
								Each line specifies a pattern. Patterns cannot be empty.`,
			},
			"enable_merge_whitelist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a merge whitelist is used.",
			},
			"merge_whitelist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allow only allowlisted users to merge pull requests into this branch.",
			},
			"merge_whitelist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allow only allowlisted teams to merge pull requests into this branch.",
			},
			"block_merge_on_rejected_reviews": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Merging will not be possible when changes are
								requested by official reviewers, even if there are enough
								approvals.`,
			},
			"block_merge_on_official_review_requests": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Merging will not be possible when it has official
								review requests, even if there are enough approvals.`,
			},
			"block_merge_on_outdated_branch": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Merging will not be possible when head branch is behind base branch.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
		},
		Description: "This resource allows you to create and manage branch protections for repositories.",
	}
}
