package gitea

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {

	// The actual provider
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITEA_TOKEN", nil),
				Description: descriptions["token"],
				ConflictsWith: []string{
					"username",
				},
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITEA_USERNAME", nil),
				Description: descriptions["username"],
				ConflictsWith: []string{
					"token",
				},
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITEA_PASSWORD", nil),
				Description: descriptions["password"],
				ConflictsWith: []string{
					"token",
				},
			},
			"base_url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("GITEA_BASE_URL", ""),
				Description:  descriptions["base_url"],
				ValidateFunc: validateAPIURLVersion,
			},
			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["cacert_file"],
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["insecure"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"gitea_user":                              dataSourceGiteaUser(),
			"gitea_org":                               dataSourceGiteaOrg(),
			"gitea_team":                              dataSourceGiteaTeam(),
			"gitea_teams":                             dataSourceGiteaTeams(),
			"gitea_team_members":                      dataSourceGiteaTeamMembers(),
			"gitea_repo":                              dataSourceGiteaRepo(),
			"gitea_repository_file":                   dataSourceGiteaRepositoryFile(),
			"gitea_repository_files":                  dataSourceGiteaRepositoryFiles(),
			"gitea_package_versions":                  dataSourceGiteaPackageVersions(),
			"gitea_repository_issue_config":           dataSourceGiteaRepositoryIssueConfig(),
			"gitea_repository_licenses":               dataSourceGiteaRepositoryLicenses(),
			"gitea_repository_signing_key":            dataSourceGiteaRepositorySigningKey(),
			"gitea_repository_subscribers":            dataSourceGiteaRepositorySubscribers(),
			"gitea_team_repository":                   dataSourceGiteaTeamRepository(),
			"gitea_pull_request_by_base_head":         dataSourceGiteaPullRequestByBaseHead(),
			"gitea_repository_actions_workflows":      dataSourceGiteaRepositoryActionsWorkflows(),
			"gitea_repository_actions_artifacts":      dataSourceGiteaRepositoryActionsArtifacts(),
			"gitea_issue_attachments":                 dataSourceGiteaIssueAttachments(),
			"gitea_issue_comment_attachments":         dataSourceGiteaIssueCommentAttachments(),
			"gitea_actions_runners":                   dataSourceGiteaActionsRunners(),
			"gitea_actions_runs":                      dataSourceGiteaActionsRuns(),
			"gitea_actions_jobs":                      dataSourceGiteaActionsJobs(),
			"gitea_actions_runner_registration_token": dataSourceGiteaActionsRunnerRegistrationToken(),
			"gitea_repository_actions_artifact":       dataSourceGiteaRepositoryActionsArtifact(),
			"gitea_repository_webhook":                dataSourceGiteaRepositoryWebhook(),
			"gitea_repositories":                      dataSourceGiteaRepos(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"gitea_org":                               resourceGiteaOrg(),
			"gitea_user":                              resourceGiteaUser(),
			"gitea_oauth2_app":                        resourceGiteaOauthApp(),
			"gitea_repository":                        resourceGiteaRepository(),
			"gitea_fork":                              resourceGiteaFork(),
			"gitea_public_key":                        resourceGiteaPublicKey(),
			"gitea_gpg_key":                           resourceGiteaGPGKey(),
			"gitea_team":                              resourceGiteaTeam(),
			"gitea_team_membership":                   resourceGiteaTeamMembership(),
			"gitea_team_members":                      resourceGiteaTeamMembers(),
			"gitea_git_hook":                          resourceGiteaGitHook(),
			"gitea_token":                             resourceGiteaToken(),
			"gitea_repository_key":                    resourceGiteaRepositoryKey(),
			"gitea_repository_webhook":                resourceGiteaRepositoryWebhook(),
			"gitea_repository_branch_protection":      resourceGiteaRepositoryBranchProtection(),
			"gitea_repository_actions_variable":       resourceGiteaRepositoryActionsVariable(),
			"gitea_repository_actions_secret":         resourceGiteaRepositoryActionsSecret(),
			"gitea_repository_branch":                 resourceGiteaRepositoryBranch(),
			"gitea_repository_file":                   resourceGiteaRepositoryFile(),
			"gitea_repository_collaborator":           resourceGiteaRepositoryCollaborator(),
			"gitea_org_actions_variable":              resourceGiteaOrgActionsVariable(),
			"gitea_org_actions_secret":                resourceGiteaOrgActionsSecret(),
			"gitea_user_actions_variable":             resourceGiteaUserActionsVariable(),
			"gitea_user_actions_secret":               resourceGiteaUserActionsSecret(),
			"gitea_issue_attachment":                  resourceGiteaIssueAttachment(),
			"gitea_issue_comment_attachment":          resourceGiteaIssueCommentAttachment(),
			"gitea_repository_actions_workflow_state": resourceGiteaRepositoryActionsWorkflowState(),
		},

		ConfigureFunc: providerConfigure,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token":       "The application token used to connect to Gitea.",
		"username":    "Username in case of using basic auth",
		"password":    "Password in case of using basic auth",
		"base_url":    "The Gitea Base API URL",
		"cacert_file": "A file containing the ca certificate to use in case ssl certificate is not from a standard chain",
		"insecure":    "Disable SSL verification of API calls",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:      d.Get("token").(string),
		Username:   d.Get("username").(string),
		Password:   d.Get("password").(string),
		BaseURL:    d.Get("base_url").(string),
		CACertFile: d.Get("cacert_file").(string),
		Insecure:   d.Get("insecure").(bool),
	}

	return config.Client()
}

func validateAPIURLVersion(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	if strings.HasSuffix(v, "/api/v1") || strings.HasSuffix(v, "/api/v1/") {
		es = append(es, fmt.Errorf("terraform-gitea-provider base URL format is incorrect; Please leave out API Path %s", v))
	}
	if strings.Contains(v, "localhost") && strings.Contains(v, ".") {
		es = append(es, fmt.Errorf("terraform-gitea-provider base URL violates RFC 2606; Please do not define a subdomain for localhost!"))
	}
	return
}
