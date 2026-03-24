package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaTeamMembers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaTeamMembersRead,
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the team.",
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"full_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"avatar_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGiteaTeamMembersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	teamID := int64(d.Get("team_id").(int))

	members, err := collectPaginated(func(page int) ([]*gitea.User, error) {
		items, _, err := client.ListTeamMembers(teamID, gitea.ListTeamMembersOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
		return items, err
	})
	if err != nil {
		return fmt.Errorf("unable to list members for team %d: %w", teamID, err)
	}

	d.SetId(fmt.Sprintf("team:%s:members", strconv.FormatInt(teamID, 10)))
	if err := d.Set("members", flattenUsers(members)); err != nil {
		return fmt.Errorf("failed to set members: %w", err)
	}

	return nil
}
