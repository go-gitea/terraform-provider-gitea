package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	membershipTeamID   string = "team_id"
	membershipUserName string = "username"
)

func parseTeamMembershipID(id string) (teamID int, username string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		if idx := strings.LastIndex(id, "_"); idx > 0 {
			parts = []string{id[:idx], id[idx+1:]}
		}
	}
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("unexpected ID format (%q), expected <team_id>/<username>", id)
	}

	teamID64, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid team_id in ID %q: %w", id, err)
	}

	return int(teamID64), parts[1], nil
}

func resourceTeamMembershipCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	team_id := d.Get(membershipTeamID).(int)
	username := d.Get(membershipUserName).(string)

	// Create the membership
	_, err = client.AddTeamMember(int64(team_id), username)

	// What if the membership exists? Consider error messages
	// Does this do anything? Will err not be return in the end anyway
	if err != nil {
		return
	}

	err = setTeamMembershipData(team_id, username, d)

	return
}

func resourceTeamMembershipRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var resp *gitea.Response

	team_id, username, err := parseTeamMembershipID(d.Id())
	if err != nil {
		return err
	}

	// Attempt to get the user from the team. If the user is not a member of the team, this will return a 404
	_, resp, err = client.GetTeamMember(int64(team_id), username)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	// The membership does not exist in Gitea
	if resp != nil && resp.StatusCode == 404 {
		// No ID in the resource indicates that it does not exist
		d.SetId("")
		return nil
	}

	err = setTeamMembershipData(team_id, username, d)

	return
}

func resourceTeamMembershipDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	team_id, username, err := parseTeamMembershipID(d.Id())
	if err != nil {
		return err
	}

	// Delete the membership
	var resp *gitea.Response
	resp, err = client.RemoveTeamMember(int64(team_id), username)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil
		}
		return err
	}

	return
}

func setTeamMembershipData(team_id int, username string, d *schema.ResourceData) (err error) {
	// This can't be team or usename only as that would not be unique since the
	// team can have multiple members and the user can have multiple memberships.
	d.SetId(fmt.Sprintf("%d/%s", team_id, username))
	d.Set(membershipTeamID, team_id)
	d.Set(membershipUserName, username)

	return
}

func resourceGiteaTeamMembership() *schema.Resource {
	return &schema.Resource{
		Read:   resourceTeamMembershipRead,
		Create: resourceTeamMembershipCreate,
		Delete: resourceTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				teamID, username, err := parseTeamMembershipID(d.Id())
				if err != nil {
					return nil, err
				}
				if err := d.Set(membershipTeamID, teamID); err != nil {
					return nil, err
				}
				if err := d.Set(membershipUserName, username); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%d/%s", teamID, username))
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the team.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the team member.",
			},
		},
		Description: "`gitea_team_membership` manages a single user's membership of a single team.",
	}
}
