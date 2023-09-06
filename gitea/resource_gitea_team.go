package gitea

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TeamName                string = "name"
	TeamOrg                 string = "organisation"
	TeamDescription         string = "description"
	TeamPermissions         string = "permission"
	TeamCreateRepoFlag      string = "can_create_repos"
	TeamIncludeAllReposFlag string = "include_all_repositories"
	TeamUnits               string = "units"
	TeamMembers             string = "members"
	TeamRepositories        string = "repositories"
)

func resourceTeamRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response
	var team *gitea.Team

	team, resp, err = client.GetTeam(id)

	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	err = setTeamResourceData(team, d, meta)

	return
}

func resourceTeamCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var team *gitea.Team
	var units []gitea.RepoUnitType

	if strings.Contains(d.Get(TeamUnits).(string), "repo.code") {
		units = append(units, gitea.RepoUnitCode)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.issues") {
		units = append(units, gitea.RepoUnitIssues)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.ext_issues") {
		units = append(units, gitea.RepoUnitExtIssues)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.wiki") {
		units = append(units, gitea.RepoUnitWiki)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.pulls") {
		units = append(units, gitea.RepoUnitPulls)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.releases") {
		units = append(units, gitea.RepoUnitReleases)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.ext_wiki") {
		units = append(units, gitea.RepoUnitExtWiki)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.projects") {
		units = append(units, gitea.RepoUnitProjects)
	}

	includeAllRepos := d.Get(TeamIncludeAllReposFlag).(bool)

	opts := gitea.CreateTeamOption{
		Name:                    d.Get(TeamName).(string),
		Description:             d.Get(TeamDescription).(string),
		Permission:              gitea.AccessMode(d.Get(TeamPermissions).(string)),
		CanCreateOrgRepo:        d.Get(TeamCreateRepoFlag).(bool),
		IncludesAllRepositories: includeAllRepos,
		Units:                   units,
	}

	team, _, err = client.CreateTeam(d.Get(TeamOrg).(string), opts)

	if err != nil {
		return
	}

	users := d.Get(TeamMembers).([]interface{})

	for _, user := range users {
		if user != "" {
			_, err = client.AddTeamMember(team.ID, user.(string))
			if err != nil {
				return err
			}
		}
	}

	if !includeAllRepos {
		err = setTeamRepositories(team, d, meta, false)
		if err != nil {
			return err
		}
	}

	err = setTeamResourceData(team, d, meta)

	return
}

func resourceTeamUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response
	var team *gitea.Team

	team, resp, err = client.GetTeam(id)

	if err != nil {
		if resp.StatusCode == 404 {
			resourceTeamCreate(d, meta)
		} else {
			return err
		}
	}

	description := d.Get(TeamDescription).(string)
	canCreateRepo := d.Get(TeamCreateRepoFlag).(bool)
	includeAllRepos := d.Get(TeamIncludeAllReposFlag).(bool)

	var units []gitea.RepoUnitType

	if strings.Contains(d.Get(TeamUnits).(string), "repo.code") {
		units = append(units, gitea.RepoUnitCode)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.issues") {
		units = append(units, gitea.RepoUnitIssues)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.ext_issues") {
		units = append(units, gitea.RepoUnitExtIssues)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.wiki") {
		units = append(units, gitea.RepoUnitWiki)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.pulls") {
		units = append(units, gitea.RepoUnitPulls)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.releases") {
		units = append(units, gitea.RepoUnitReleases)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.ext_wiki") {
		units = append(units, gitea.RepoUnitExtWiki)
	}
	if strings.Contains(d.Get(TeamUnits).(string), "repo.projects") {
		units = append(units, gitea.RepoUnitProjects)
	}

	opts := gitea.EditTeamOption{
		Name:                    d.Get(TeamName).(string),
		Description:             &description,
		Permission:              gitea.AccessMode(d.Get(TeamPermissions).(string)),
		CanCreateOrgRepo:        &canCreateRepo,
		IncludesAllRepositories: &includeAllRepos,
		Units:                   units,
	}

	resp, err = client.EditTeam(id, opts)

	if err != nil {
		return err
	}

	users := d.Get(TeamMembers).([]interface{})

	for _, user := range users {
		if user != "" {
			_, err = client.AddTeamMember(team.ID, user.(string))
			if err != nil {
				return err
			}
		}
	}

	if !includeAllRepos {
		err = setTeamRepositories(team, d, meta, true)
		if err != nil {
			return err
		}
	}

	team, _, _ = client.GetTeam(id)

	err = setTeamResourceData(team, d, meta)

	return
}

func resourceTeamDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response

	resp, err = client.DeleteTeam(id)

	if err != nil {
		if resp.StatusCode == 404 {
			return
		} else {
			return err
		}
	}

	return
}

func setTeamResourceData(team *gitea.Team, d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)
	if err := client.CheckServerVersionConstraint(">= 1.19.4"); err != nil {
		d.Set(TeamOrg, d.Get(TeamOrg).(string))
	} else {
		d.Set(TeamOrg, team.Organization.UserName)
	}
	d.SetId(fmt.Sprintf("%d", team.ID))
	d.Set(TeamCreateRepoFlag, team.CanCreateOrgRepo)
	d.Set(TeamDescription, team.Description)
	d.Set(TeamName, team.Name)
	d.Set(TeamPermissions, string(team.Permission))
	d.Set(TeamIncludeAllReposFlag, team.IncludesAllRepositories)
	d.Set(TeamUnits, d.Get(TeamUnits).(string))
	d.Set(TeamMembers, d.Get(TeamMembers))
	d.Set(TeamRepositories, d.Get(TeamRepositories))
	return
}

func resourceGiteaTeam() *schema.Resource {
	return &schema.Resource{
		Read:   resourceTeamRead,
		Create: resourceTeamCreate,
		Update: resourceTeamUpdate,
		Delete: resourceTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Team",
			},
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organisation which this Team is part of.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "Description of the Team",
			},
			"permission": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
				Description: "Permissions associated with this Team\n" +
					"Can be `none`, `read`, `write`, `admin` or `owner`",
			},
			"can_create_repos": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     true,
				Description: "Flag if the Teams members should be able to create Rpositories in the Organisation",
			},
			"include_all_repositories": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     true,
				Description: "Flag if the Teams members should have access to all Repositories in the Organisation",
			},
			"units": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "[repo.code, repo.issues, repo.ext_issues, repo.wiki, repo.pulls, repo.releases, repo.projects, repo.ext_wiki]",
				Description: "List of types of Repositories that should be allowed to be created from Team members.\n" +
					"Can be `repo.code`, `repo.issues`, `repo.ext_issues`, `repo.wiki`, `repo.pulls`, `repo.releases`, `repo.projects` and/or `repo.ext_wiki`",
			},
			"members": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Required:    false,
				Computed:    true,
				Description: "List of Users that should be part of this team",
			},
			"repositories": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Required:    false,
				Computed:    true,
				Description: "List of Repositories that should be part of this team",
			},
		},
		Description: "`gitea_team` manages Team that are part of an organisation.",
	}
}

func setTeamRepositories(team *gitea.Team, d *schema.ResourceData, meta interface{}, update bool) (err error) {
	client := meta.(*gitea.Client)

	org := d.Get(TeamOrg).(string)

	repositories := make(map[string]bool)
	for _, repo := range d.Get(TeamRepositories).([]interface{}) {
		if repo != "" {
			repositories[repo.(string)] = true
		}
	}

	if update {
		page := 1

		for {
			var existingRepositories []*gitea.Repository
			existingRepositories, _, err = client.ListTeamRepositories(team.ID, gitea.ListTeamRepositoriesOptions{
				ListOptions: gitea.ListOptions{
					Page:     page,
					PageSize: 50,
				},
			})
			if err != nil {
				return errors.New(fmt.Sprintf("[ERROR] Error listeng team repositories: %s", err))
			}
			if len(existingRepositories) == 0 {
				break
			}

			for _, exr := range existingRepositories {
				_, exists := repositories[exr.Name]
				if exists {
					repositories[exr.Name] = false
				} else {
					_, err = client.RemoveTeamRepository(team.ID, org, exr.Name)
					if err != nil {
						return errors.New(fmt.Sprintf("[ERROR] Error removing team repository %q: %s", exr.Name, err))
					}
				}
			}

			page += 1
		}
	}

	for repo, flag := range repositories {
		if flag {
			_, err = client.AddTeamRepository(team.ID, org, repo)
			if err != nil {
				return errors.New(fmt.Sprintf("[ERROR] Error adding team repository %q: %s", repo, err))
			}
		}
	}

	return
}
