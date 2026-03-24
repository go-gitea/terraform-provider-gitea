package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaTeams() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaTeamsRead,
		Schema: map[string]*schema.Schema{
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organisation whose teams to list.",
			},
			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"permission": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"can_create_repos": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_all_repositories": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"units": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceGiteaTeamsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	org := d.Get("organisation").(string)

	teams, err := collectPaginated(func(page int) ([]*gitea.Team, error) {
		items, _, err := client.ListOrgTeams(org, gitea.ListTeamsOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
		return items, err
	})
	if err != nil {
		return fmt.Errorf("unable to list teams for org %q: %w", org, err)
	}

	result := make([]interface{}, 0, len(teams))
	for _, team := range teams {
		units := make([]string, len(team.Units))
		for i, u := range team.Units {
			units[i] = string(u)
		}
		result = append(result, map[string]interface{}{
			"id":                       int(team.ID),
			"name":                     team.Name,
			"description":              team.Description,
			"permission":               string(team.Permission),
			"can_create_repos":         team.CanCreateOrgRepo,
			"include_all_repositories": team.IncludesAllRepositories,
			"units":                    units,
		})
	}

	d.SetId(fmt.Sprintf("org:%s:teams", org))
	if err := d.Set("teams", result); err != nil {
		return fmt.Errorf("failed to set teams: %w", err)
	}

	return nil
}
