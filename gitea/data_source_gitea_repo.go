package gitea

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepo() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepoRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"fork": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mirror": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clone_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stars": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"forks": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"watchers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"open_issue_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permission_admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"permission_push": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"permission_pull": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceGiteaRepoRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	usernameData, usernameOk := d.GetOk("username")
	if !usernameOk {
		return fmt.Errorf("name of repo owner must be passed")
	}
	username := strings.ToLower(usernameData.(string))

	nameData, nameOk := d.GetOk("name")
	if !nameOk {
		return fmt.Errorf("name of repo must be passed")
	}
	name := strings.ToLower(nameData.(string))

	repo, _, err := client.GetRepo(username, name)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", repo.ID))
	d.Set("name", repo.Name)
	d.Set("description", repo.Description)
	d.Set("full_name", repo.FullName)
	d.Set("description", repo.Description)
	d.Set("private", repo.Private)
	d.Set("fork", repo.Fork)
	d.Set("mirror", repo.Mirror)
	d.Set("size", repo.Size)
	d.Set("html_url", repo.HTMLURL)
	d.Set("ssh_url", repo.SSHURL)
	d.Set("clone_url", repo.CloneURL)
	d.Set("website", repo.Website)
	d.Set("stars", repo.Stars)
	d.Set("forks", repo.Forks)
	d.Set("watchers", repo.Watchers)
	d.Set("open_issue_count", repo.OpenIssues)
	d.Set("default_branch", repo.DefaultBranch)
	d.Set("created", repo.Created)
	d.Set("updated", repo.Updated)
	d.Set("permission_admin", repo.Permissions.Admin)
	d.Set("permission_push", repo.Permissions.Push)
	d.Set("permission_pull", repo.Permissions.Pull)
	return nil
}
