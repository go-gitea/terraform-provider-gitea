package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	membersTeamID 	   string = "team_id"
	membersTeamMembers string = "members"
)

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

func resourceTeamMembersCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)
	team_id := d.Get(membersTeamID).(int)

	var memberNames []string 

	// What if team already has member?
	// What if user is already in the team?
	// What if user does not exist?
	
	// Add members to the team
	for _, name := range d.Get(membersTeamMembers).(*schema.Set).List() {
		_ , err = client.AddTeamMember(int64(team_id), name.(string))
		if err != nil {
			return err
		}
		// Update list of usernames of the team members
		memberNames = append(memberNames, name.(string))
	}

	err = setTeamMembersData(team_id, memberNames, d)

	return
}

func resourceTeamMembersRead(d *schema.ResourceData, meta interface{}) (err error) {
	team_id := d.Get(membersTeamID).(int)

	memberNames, err := getTeamMembers(team_id, meta)
	if err != nil {
		return err
	}

	err = setTeamMembersData(team_id, memberNames, d)

	return
}

func resourceTeamMembersDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)
	team_id := d.Get(membersTeamID).(int)

	var memberNames []string

	memberNames , err = getTeamMembers(team_id, meta)
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
			StateContext: schema.ImportStatePassthroughContext,
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
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				ForceNew:	 true,
				Description: "The user names of the members of the team.",
			},

		},
		Description: "`gitea_team_members` manages all members of a single team. This resource will be recreated on member changes.",
	}
}
