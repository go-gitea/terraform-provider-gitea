package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	forkOwner        string = "owner"
	forkRepo         string = "repo"
	forkOrganization string = "organization"
)

func resourceForkCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var opts gitea.CreateForkOption
	var org string
	org = d.Get(forkOrganization).(string)
	if org != "" {
		opts.Organization = &org
	}

	repo, _, err := client.CreateFork(d.Get(forkOwner).(string),
		d.Get(forkRepo).(string),
		opts)
	if err == nil {
		err = setForkResourceData(repo, d)
	}
	return err
}

func resourceForkRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	var resp *gitea.Response

	if err != nil {
		return err
	}

	repo, resp, err := client.GetRepoByID(id)

	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	err = setForkResourceData(repo, d)

	return
}

func resourceForkDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	if err != nil {
		return err
	}

	repo, _, err := client.GetRepoByID(id)
	var resp *gitea.Response

	resp, err = client.DeleteRepo(repo.Owner.UserName, repo.Name)

	if err != nil {
		if resp.StatusCode == 404 {
			return
		} else {
			return err
		}
	}

	return
}

func setForkResourceData(repo *gitea.Repository, d *schema.ResourceData) (err error) {

	d.SetId(fmt.Sprintf("%d", repo.ID))

	return
}

func resourceGiteaFork() *schema.Resource {
	return &schema.Resource{
		Read:   resourceForkRead,
		Create: resourceForkCreate,
		Delete: resourceForkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The owner or owning organization of the repository to fork",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository to fork",
			},
			"organization": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				ForceNew:    true,
				Description: "The organization that owns the forked repo",
			},
		},
		Description: "`gitea_fork` manages repository fork to the current user or an organisation\n" +
			"Forking a repository to a dedicated user is currently unsupported\n" +
			"Creating a fork using this resource without an organisation will create the fork in the executors name",
	}
}
