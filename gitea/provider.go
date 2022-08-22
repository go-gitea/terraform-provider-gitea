package gitea

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

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
			"gitea_user": dataSourceGiteaUser(),
			"gitea_org":  dataSourceGiteaOrg(),
			// "gitea_team":   dataSourceGiteaTeam(),
			// "gitea_teams":  dataSourceGiteaTeams(),
			// "gitea_team_members":  dataSourceGiteaTeamMembers(),
			"gitea_repo": dataSourceGiteaRepo(),
			// "gitea_repos":  dataSourceGiteaRepos(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"gitea_org": resourceGiteaOrg(),
			// "gitea_team": resourceGiteaTeam(),
			// "gitea_repo": resourceGiteaRepo(),
			"gitea_user":       resourceGiteaUser(),
			"gitea_oauth2_app": resourceGiteaOauthApp(),
			"gitea_repository": resourceGiteaRepository(),
			"gitea_public_key": resourceGiteaPublicKey(),
			"gitea_team":       resourceGiteaTeam(),
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
	return
}
