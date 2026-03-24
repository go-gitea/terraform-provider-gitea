package gitea

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	membersTeamID      string = "team_id"
	membersTeamMembers string = "members"
)

func parseTeamMembersID(id string) (int, error) {
	teamID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("unexpected ID format (%q), expected <team_id>", id)
	}
	return teamID, nil
}

func getTeamMembers(team_id int, meta interface{}) (membersNames []string, err error) {
	client := meta.(*gitea.Client)

	var memberNames []string
	var members []*gitea.User

	// Get all pages of users
	page := 1
	for {
		// Set options for current page
		opts := gitea.ListTeamMembersOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		}

		// Get page of team members
		members, _, err = client.ListTeamMembers(int64(team_id), opts)
		if err != nil {
			return nil, err
		}

		// If no members were returned, we are done
		if len(members) == 0 {
			break
		}

		// Update list of usernames with data from current page
		for _, m := range members {
			memberNames = append(memberNames, m.UserName)
		}

		// Next page
		page += 1
	}

	return memberNames, nil
}

func teamMembersDiff(current, desired []string) (toAdd, toRemove []string) {
	currentSet := make(map[string]struct{}, len(current))
	desiredSet := make(map[string]struct{}, len(desired))

	for _, name := range current {
		currentSet[name] = struct{}{}
	}
	for _, name := range desired {
		desiredSet[name] = struct{}{}
	}

	for _, name := range desired {
		if _, exists := currentSet[name]; !exists {
			toAdd = append(toAdd, name)
		}
	}

	for _, name := range current {
		if _, exists := desiredSet[name]; !exists {
			toRemove = append(toRemove, name)
		}
	}

	return toAdd, toRemove
}

func resourceTeamMembersCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)
	team_id := d.Get(membersTeamID).(int)

	desiredMembers := make([]string, 0)
	for _, name := range d.Get(membersTeamMembers).(*schema.Set).List() {
		desiredMembers = append(desiredMembers, name.(string))
	}

	currentMembers, err := getTeamMembers(team_id, meta)
	if err != nil {
		return err
	}

	toAdd, toRemove := teamMembersDiff(currentMembers, desiredMembers)

	for _, username := range toRemove {
		_, err = client.RemoveTeamMember(int64(team_id), username)
		if err != nil {
			return err
		}
	}

	for _, username := range toAdd {
		_, err = client.AddTeamMember(int64(team_id), username)
		if err != nil {
			return err
		}
	}

	memberNames, err := getTeamMembers(team_id, meta)
	if err != nil {
		return err
	}

	err = setTeamMembersData(team_id, memberNames, d)

	return
}

func resourceTeamMembersRead(d *schema.ResourceData, meta interface{}) (err error) {
	team_id, err := parseTeamMembersID(d.Id())
	if err != nil {
		return err
	}

	memberNames, err := getTeamMembers(team_id, meta)
	if err != nil {
		return err
	}

	err = setTeamMembersData(team_id, memberNames, d)

	return
}

func resourceTeamMembersDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)
	team_id, err := parseTeamMembersID(d.Id())
	if err != nil {
		return err
	}

	var memberNames []string

	memberNames, err = getTeamMembers(team_id, meta)
	if err != nil {
		return err
	}

	// Delete all memberships
	for _, username := range memberNames {
		_, err = client.RemoveTeamMember(int64(team_id), username)
		if err != nil {
			return err
		}
	}

	return
}

func setTeamMembersData(team_id int, memberNames []string, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d", team_id))
	d.Set(membersTeamID, team_id)
	d.Set(membersTeamMembers, memberNames)

	return
}

func resourceGiteaTeamMembers() *schema.Resource {
	return &schema.Resource{
		Read:   resourceTeamMembersRead,
		Create: resourceTeamMembersCreate,
		Delete: resourceTeamMembersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				teamID, err := parseTeamMembersID(d.Id())
				if err != nil {
					return nil, err
				}
				if err := d.Set(membersTeamID, teamID); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%d", teamID))
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
			"members": {
				// TypeSet is better than TypeList because
				// reordering the members will not trigger recreation
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				ForceNew:    true,
				Description: "The user names of the members of the team.",
			},
		},
		Description: "`gitea_team_members` manages all members of a single team. This resource will be recreated on member changes.",
	}
}
