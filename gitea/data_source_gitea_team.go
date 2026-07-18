package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaTeamRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
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
			"units_map": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Map of repository units to their permissions",
			},
		},
	}
}

func dataSourceGiteaTeamRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	id := int64(d.Get("id").(int))

	team, _, err := client.GetTeam(id)
	if err != nil {
		return fmt.Errorf("unable to retrieve team %d: %w", id, err)
	}

	d.SetId(strconv.FormatInt(team.ID, 10))
	d.Set("name", team.Name)
	d.Set("description", team.Description)
	d.Set("permission", string(team.Permission))
	d.Set("can_create_repos", team.CanCreateOrgRepo)
	d.Set("include_all_repositories", team.IncludesAllRepositories)

	if team.Organization != nil {
		d.Set("organisation", team.Organization.UserName)
	}

	units := make([]string, len(team.Units))
	for i, u := range team.Units {
		units[i] = string(u)
	}
	if err := d.Set("units", units); err != nil {
		return fmt.Errorf("failed to set units: %w", err)
	}

	if err := d.Set("units_map", team.UnitsMap); err != nil {
		return fmt.Errorf("failed to set units_map: %w", err)
	}

	return nil
}
