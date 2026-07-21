package gitea

import (
	"reflect"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSetRepositoryWebhookDataUsesHookFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryWebhook().Schema, map[string]interface{}{
		"username":             "stale-owner",
		"name":                 "stale-repo",
		"type":                 "stale-type",
		"url":                  "https://stale.example.com",
		"content_type":         "form",
		"events":               []interface{}{"old"},
		"branch_filter":        "stale-branch",
		"active":               false,
		"secret":               "sticky-secret",
		"authorization_header": "old-auth",
	})

	hook := &gitea.Hook{
		ID:                  17,
		Type:                "discord",
		Config:              map[string]string{"url": "https://api.example.com/hook", "content_type": "json"},
		Events:              []string{"push", "pull_request"},
		BranchFilter:        "feature/*",
		Active:              true,
		AuthorizationHeader: "Bearer updated",
	}

	if err := setRepositoryWebhookData(hook, d); err != nil {
		t.Fatalf("setRepositoryWebhookData returned error: %v", err)
	}

	if got := d.Id(); got != "17" {
		t.Fatalf("expected id 17, got %q", got)
	}
	if got := d.Get(repoWebhookType).(string); got != "discord" {
		t.Fatalf("expected type discord, got %q", got)
	}
	if got := d.Get(repoWebhookUrl).(string); got != "https://api.example.com/hook" {
		t.Fatalf("expected url from hook config, got %q", got)
	}
	if got := d.Get(repoWebhookContentType).(string); got != "json" {
		t.Fatalf("expected content_type json, got %q", got)
	}
	if got := resourceDataStringList(d, repoWebhookEvents); !reflect.DeepEqual(got, []string{"push", "pull_request"}) {
		t.Fatalf("expected events from hook, got %#v", got)
	}
	if got := d.Get(repoWebhookBranchFilter).(string); got != "feature/*" {
		t.Fatalf("expected branch_filter from hook, got %q", got)
	}
	if got := d.Get(repoWebhookActive).(bool); !got {
		t.Fatalf("expected active to be true")
	}
	if got := d.Get(repoWebhookSecret).(string); got != "sticky-secret" {
		t.Fatalf("expected secret to stay sticky, got %q", got)
	}
	if got := d.Get(repoWebhookAuthorizationHeader).(string); got != "Bearer updated" {
		t.Fatalf("expected authorization header from hook, got %q", got)
	}
}

func TestSetRepoResourceDataUsesRepositoryFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepository().Schema, map[string]interface{}{
		"username":                    "stale-owner",
		"name":                        "stale-name",
		"description":                 "stale-description",
		"private":                     false,
		"repo_template":               false,
		"mirror":                      false,
		"has_issues":                  false,
		"has_wiki":                    false,
		"has_pull_requests":           false,
		"has_projects":                false,
		"ignore_whitespace_conflicts": false,
		"allow_merge_commits":         false,
		"allow_rebase":                false,
		"allow_rebase_explicit":       false,
		"allow_squash_merge":          false,
		"archived":                    false,
		"allow_manual_merge":          false,
		"autodetect_manual_merge":     false,
		"migration_mirror_interval":   "0s",
	})

	repo := &gitea.Repository{
		ID:                        88,
		Owner:                     &gitea.User{UserName: "api-owner"},
		Name:                      "api-name",
		Description:               "api-description",
		FullName:                  "api-owner/api-name",
		Private:                   true,
		Fork:                      true,
		Template:                  true,
		Mirror:                    true,
		Size:                      12,
		HTMLURL:                   "https://gitea.example.com/api-owner/api-name",
		SSHURL:                    "ssh://gitea.example.com/api-owner/api-name.git",
		CloneURL:                  "https://gitea.example.com/api-owner/api-name.git",
		Website:                   "https://example.com",
		Stars:                     4,
		Forks:                     5,
		Watchers:                  6,
		OpenIssues:                7,
		DefaultBranch:             "main",
		Created:                   time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC),
		Updated:                   time.Date(2024, time.March, 4, 5, 6, 7, 0, time.UTC),
		Permissions:               &gitea.Permission{Admin: true, Push: false, Pull: true},
		HasIssues:                 true,
		HasWiki:                   true,
		HasPullRequests:           true,
		HasProjects:               true,
		IgnoreWhitespaceConflicts: true,
		AllowMerge:                true,
		AllowRebase:               false,
		AllowRebaseMerge:          true,
		AllowSquash:               false,
		Archived:                  true,
		MirrorInterval:            "8h0m0s",
		DefaultMergeStyle:         gitea.MergeStyleRebaseMerge,
	}

	if err := setRepoResourceData(repo, d); err != nil {
		t.Fatalf("setRepoResourceData returned error: %v", err)
	}

	if got := d.Get(repoDefaultMergeStyle).(string); got != "rebase-merge" {
		t.Fatalf("expected default_merge_style rebase-merge, got %q", got)
	}

	if got := d.Get(repoOwner).(string); got != "api-owner" {
		t.Fatalf("expected owner api-owner, got %q", got)
	}
	if got := d.Get(repoName).(string); got != "api-name" {
		t.Fatalf("expected repo name api-name, got %q", got)
	}
	if got := d.Get(repoDescription).(string); got != "api-description" {
		t.Fatalf("expected description from repo, got %q", got)
	}
	if got := d.Get(repoTemplate).(bool); !got {
		t.Fatalf("expected template to be true")
	}
	if got := d.Get(repoMirror).(bool); !got {
		t.Fatalf("expected mirror to be true")
	}
	if got := d.Get(repoIssues).(bool); !got {
		t.Fatalf("expected has_issues to be true")
	}
	if got := d.Get(repoWiki).(bool); !got {
		t.Fatalf("expected has_wiki to be true")
	}
	if got := d.Get(repoPrs).(bool); !got {
		t.Fatalf("expected has_pull_requests to be true")
	}
	if got := d.Get(repoProjects).(bool); !got {
		t.Fatalf("expected has_projects to be true")
	}
	if got := d.Get(repoIgnoreWhitespace).(bool); !got {
		t.Fatalf("expected ignore_whitespace_conflicts to be true")
	}
	if got := d.Get(repoAllowMerge).(bool); !got {
		t.Fatalf("expected allow_merge_commits to be true")
	}
	if got := d.Get(repoAllowRebase).(bool); got {
		t.Fatalf("expected allow_rebase to be false")
	}
	if got := d.Get(repoAllowRebaseMerge).(bool); !got {
		t.Fatalf("expected allow_rebase_explicit to be true")
	}
	if got := d.Get(repoAllowSquash).(bool); got {
		t.Fatalf("expected allow_squash_merge to be false")
	}
	if got := d.Get(repoArchived).(bool); !got {
		t.Fatalf("expected archived to be true")
	}
	if got := d.Get(migrationMirrorInterval).(string); got != "8h0m0s" {
		t.Fatalf("expected mirror interval from repo, got %q", got)
	}
	if got := d.Get("permission_admin").(bool); !got {
		t.Fatalf("expected permission_admin to be true")
	}
	if got := d.Get("permission_pull").(bool); !got {
		t.Fatalf("expected permission_pull to be true")
	}
	if got := d.Get(repoAllowManualMerge).(bool); got {
		t.Fatalf("expected allow_manual_merge to remain sticky false")
	}
	if got := d.Get(repoAutodetectManualMerge).(bool); got {
		t.Fatalf("expected autodetect_manual_merge to remain sticky false")
	}
}

func TestSetTeamResourceDataUsesTeamFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaTeam().Schema, map[string]interface{}{
		"organisation":             "stale-org",
		"name":                     "stale-team",
		"description":              "stale-description",
		"permission":               "read",
		"can_create_repos":         true,
		"include_all_repositories": false,
		"units":                    "stale-units",
		"units_map":                map[string]interface{}{"repo.code": "read"},
		"repositories":             []interface{}{"stale"},
	})

	team := &gitea.Team{
		ID:                      21,
		Name:                    "api-team",
		Description:             "api-description",
		Organization:            &gitea.Organization{UserName: "api-org"},
		Permission:              gitea.AccessModeAdmin,
		CanCreateOrgRepo:        false,
		IncludesAllRepositories: true,
		Units: []gitea.RepoUnitType{
			gitea.RepoUnitCode,
			gitea.RepoUnitProjects,
		},
		UnitsMap: map[string]string{
			"repo.code": "write",
		},
	}

	if err := setTeamResourceData(team, []string{"zeta", "alpha"}, d); err != nil {
		t.Fatalf("setTeamResourceData returned error: %v", err)
	}

	if got := d.Get(TeamOrg).(string); got != "api-org" {
		t.Fatalf("expected organisation api-org, got %q", got)
	}
	if got := d.Get(TeamName).(string); got != "api-team" {
		t.Fatalf("expected team name api-team, got %q", got)
	}
	if got := d.Get(TeamDescription).(string); got != "api-description" {
		t.Fatalf("expected description from team, got %q", got)
	}
	if got := d.Get(TeamPermissions).(string); got != string(gitea.AccessModeAdmin) {
		t.Fatalf("expected permission admin, got %q", got)
	}
	if got := d.Get(TeamCreateRepoFlag).(bool); got {
		t.Fatalf("expected can_create_repos to be false")
	}
	if got := d.Get(TeamIncludeAllReposFlag).(bool); !got {
		t.Fatalf("expected include_all_repositories to be true")
	}
	if got := d.Get(TeamUnits).(string); got != "[repo.code repo.projects]" {
		t.Fatalf("expected units from team, got %q", got)
	}
	gotUnitsMap := d.Get("units_map").(map[string]interface{})
	if got := gotUnitsMap["repo.code"].(string); got != "write" {
		t.Fatalf("expected units_map['repo.code'] to be 'write', got %q", got)
	}
	if got := resourceDataStringList(d, TeamRepositories); !reflect.DeepEqual(got, []string{"alpha", "zeta"}) {
		t.Fatalf("expected repositories from API and sorted, got %#v", got)
	}
}

func TestSetUserResourceDataUsesUserFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaUser().Schema, map[string]interface{}{
		"login_name":                "stale-login",
		"email":                     "stale@example.com",
		"full_name":                 "stale full",
		"must_change_password":      true,
		"send_notification":         false,
		"visibility":                "public",
		"description":               "stale-description",
		"location":                  "stale-location",
		"active":                    false,
		"allow_git_hook":            true,
		"allow_import_local":        true,
		"max_repo_creation":         7,
		"prohibit_login":            false,
		"allow_create_organization": true,
		"restricted":                false,
		"force_password_change":     true,
	})

	user := &gitea.User{
		ID:            99,
		UserName:      "api-user",
		LoginName:     "api-login",
		FullName:      "API User",
		Email:         "api@example.com",
		AvatarURL:     "https://gitea.example.com/avatar.png",
		Language:      "en-US",
		IsAdmin:       true,
		LastLogin:     time.Date(2024, time.February, 3, 4, 5, 6, 0, time.UTC),
		Created:       time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC),
		Restricted:    true,
		IsActive:      true,
		ProhibitLogin: true,
		Location:      "Toronto",
		Description:   "API description",
		Visibility:    gitea.VisibleTypePrivate,
	}

	if err := setUserResourceData(user, d); err != nil {
		t.Fatalf("setUserResourceData returned error: %v", err)
	}

	if got := d.Get(userName).(string); got != "api-user" {
		t.Fatalf("expected username api-user, got %q", got)
	}
	if got := d.Get(userLoginName).(string); got != "api-login" {
		t.Fatalf("expected login_name api-login, got %q", got)
	}
	if got := d.Get(userEmail).(string); got != "api@example.com" {
		t.Fatalf("expected email api@example.com, got %q", got)
	}
	if got := d.Get(userFullName).(string); got != "API User" {
		t.Fatalf("expected full_name API User, got %q", got)
	}
	if got := d.Get(userDescription).(string); got != "API description" {
		t.Fatalf("expected description from user, got %q", got)
	}
	if got := d.Get(userLocation).(string); got != "Toronto" {
		t.Fatalf("expected location Toronto, got %q", got)
	}
	if got := d.Get(userVisibility).(string); got != string(gitea.VisibleTypePrivate) {
		t.Fatalf("expected visibility private, got %q", got)
	}
	if got := d.Get(userActive).(bool); !got {
		t.Fatalf("expected active to be true")
	}
	if got := d.Get(userPhorbitLogin).(bool); !got {
		t.Fatalf("expected prohibit_login to be true")
	}
	if got := d.Get(userRestricted).(bool); !got {
		t.Fatalf("expected restricted to be true")
	}
	if got := d.Get(userAllowGitHook).(bool); !got {
		t.Fatalf("expected allow_git_hook to remain sticky true")
	}
	if got := d.Get(userAllowLocalImport).(bool); !got {
		t.Fatalf("expected allow_import_local to remain sticky true")
	}
	if got := d.Get(userMaxRepoCreation).(int); got != 7 {
		t.Fatalf("expected max_repo_creation to remain sticky 7, got %d", got)
	}
	if got := d.Get(userAllowCreateOrgs).(bool); !got {
		t.Fatalf("expected allow_create_organization to remain sticky true")
	}
	if got := d.Get(userForcePasswordChange).(bool); !got {
		t.Fatalf("expected force_password_change to remain sticky true")
	}
}

func resourceDataStringList(d *schema.ResourceData, key string) []string {
	raw := d.Get(key)
	values := make([]string, 0)
	switch typed := raw.(type) {
	case []interface{}:
		for _, value := range typed {
			values = append(values, value.(string))
		}
	case []string:
		values = append(values, typed...)
	}
	return values
}
