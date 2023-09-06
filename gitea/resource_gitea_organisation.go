package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	orgName                   string = "name"
	orgFullName               string = "full_name"
	orgDescription            string = "description"
	orgWebsite                string = "website"
	orgLocation               string = "location"
	orgVisibility             string = "visibility"
	RepoAdminChangeTeamAccess string = "repo_admin_change_team_access"
	orgRepos                  string = "org_repos"
)

// might come in handy if we want to stick to numeric IDs
func searchOrgByClientId(c *gitea.Client, id int64) (res *gitea.Organization, err error) {

	page := 1

	for {
		orgs, _, err := c.AdminListOrgs(gitea.AdminListOrgsOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 50,
			},
		})
		if err != nil {
			return nil, err
		}

		if len(orgs) == 0 {
			return nil, fmt.Errorf("Organisation with ID %d could not be found", id)
		}

		for _, org := range orgs {
			if org.ID == id {
				return org, nil
			}
		}

		page += 1
	}
}

func getAllOrgRepos(c *gitea.Client, orgName string) (repos []string, err error) {
	page := 1

	for {
		repoBuffer, _, err := c.ListOrgRepos(orgName, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 50,
			},
		})
		if err != nil {
			return nil, err
		}

		if len(repoBuffer) == 0 {
			return repos, nil
		}

		for _, repo := range repoBuffer {
			repos = append(repos, repo.Name)
		}

		page += 1
	}
}

func resourceOrgRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var org *gitea.Organization

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	org, err = searchOrgByClientId(client, id)

	if err != nil {
		d.SetId("")
		return nil
	}

	repos, _ := getAllOrgRepos(client, org.UserName)
	err = setOrgResourceData(org, d, &repos)

	return
}

func resourceOrgCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	opts := gitea.CreateOrgOption{
		Name:                      d.Get(orgName).(string),
		FullName:                  d.Get(orgFullName).(string),
		Description:               d.Get(orgDescription).(string),
		Website:                   d.Get(orgWebsite).(string),
		Location:                  d.Get(orgLocation).(string),
		Visibility:                gitea.VisibleType(d.Get(orgVisibility).(string)),
		RepoAdminChangeTeamAccess: d.Get(RepoAdminChangeTeamAccess).(bool),
	}

	org, _, err := client.CreateOrg(opts)
	if err != nil {
		return
	}

	repos, _ := getAllOrgRepos(client, org.UserName)
	err = setOrgResourceData(org, d, &repos)

	return
}

func resourceOrgUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var org *gitea.Organization
	var resp *gitea.Response

	org, resp, err = client.GetOrg(d.Get(orgName).(string))

	if err != nil {
		if resp.StatusCode == 404 {
			resourceOrgCreate(d, meta)
		} else {
			return err
		}
	}

	opts := gitea.EditOrgOption{
		FullName:    d.Get(orgFullName).(string),
		Description: d.Get(orgDescription).(string),
		Website:     d.Get(orgWebsite).(string),
		Location:    d.Get(orgLocation).(string),
		Visibility:  gitea.VisibleType(d.Get(orgVisibility).(string)),
	}

	client.EditOrg(d.Get(orgName).(string), opts)

	org, resp, err = client.GetOrg(d.Get(orgName).(string))

	repos, _ := getAllOrgRepos(client, org.UserName)
	err = setOrgResourceData(org, d, &repos)

	return
}

func resourceOrgDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var resp *gitea.Response

	resp, err = client.DeleteOrg(d.Get(orgName).(string))

	if err != nil {
		if resp.StatusCode == 404 {
			return
		} else {
			return err
		}
	}

	return
}

func setOrgResourceData(org *gitea.Organization, d *schema.ResourceData, repos *[]string) (err error) {
	d.SetId(fmt.Sprintf("%d", org.ID))
	d.Set("name", org.UserName)
	d.Set("full_name", org.FullName)
	d.Set("avatar_url", org.AvatarURL)
	d.Set("description", org.Description)
	d.Set("website", org.Website)
	d.Set("location", org.Location)
	d.Set("visibility", org.Visibility)
	d.Set("repos", repos)

	return
}

func resourceGiteaOrg() *schema.Resource {
	return &schema.Resource{
		Read:   resourceOrgRead,
		Create: resourceOrgCreate,
		Update: resourceOrgUpdate,
		Delete: resourceOrgDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the organisation without spaces.",
			},
			"full_name": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "The display name of the organisation. Defaults to the value of `name`.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "A description of this organisation.",
			},
			"website": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "A link to a website with more information about this organisation.",
			},
			"location": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"repo_admin_change_team_access": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Required: false,
				Computed: true,
			},
			"visibility": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "public",
				Description: "Flag is this organisation should be publicly visible or not.",
			},
			"repos": {
				Type:        schema.TypeList,
				Required:    false,
				Computed:    true,
				Description: "List of all Repositories that are part of this organisation",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
		Description: "`gitea_org` manages a gitea organisation.\n\n" +
			"Organisations are a way to group repositories and abstract permission management in a gitea instance.",
	}
}
