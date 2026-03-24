package gitea

import (
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	repoBranchName string = "name"
	repoBranchRepo string = "repository"
)

func resourceRepoBranchIdParts(d *schema.ResourceData) (hasId bool, repoId int64, branchId string, err error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return false, 0, "", nil
	}

	repoId, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, 0, "", err
	}

	branchId = parts[1]
	return true, repoId, branchId, err
}

func resourceRepoBranchRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	hasId, repoId, branchId, err := resourceRepoBranchIdParts(d)
	if err != nil {
		return err
	}
	if !hasId {
		d.SetId("")
		return nil
	}

	repo, resp, err := client.GetRepoByID(repoId)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	branch, resp, err := client.GetRepoBranch(repo.Owner.UserName, repo.Name, branchId)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	err = setRepoBranchResourceData(branch, repoId, d)

	return err
}

func resourceRepoBranchCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	repo, _, err := client.GetRepoByID(int64(d.Get(repoBranchRepo).(int)))

	if err != nil {
		return err
	}

	rb, _, err := client.CreateBranch(repo.Owner.UserName, repo.Name, gitea.CreateBranchOption{
		BranchName: d.Get(repoBranchName).(string),
	})

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d/%s", repo.ID, d.Get(repoBranchName).(string)))

	err = setRepoBranchResourceData(rb, repo.ID, d)
	return err
}

func resourceRepoBranchDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)
	hasId, repoId, branchId, err := resourceRepoBranchIdParts(d)
	if err != nil {
		return err
	}
	if !hasId {
		d.SetId("")
		return nil
	}

	repo, resp, err := client.GetRepoByID(repoId)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	client.DeleteRepoBranch(repo.Owner.UserName, repo.Name, branchId)
	return nil
}

func setRepoBranchResourceData(rb *gitea.Branch, repoId int64, d *schema.ResourceData) (err error) {
	d.Set(repoBranchName, rb.Name)
	d.Set(repoBranchRepo, repoId)
	return
}

func resourceGiteaRepositoryBranch() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepoBranchRead,
		Create: resourceRepoBranchCreate,
		Delete: resourceRepoBranchDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			repoBranchName: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the created branch",
			},
			repoBranchRepo: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the target repository",
			},
		},
		Description: "`gitea_repository_branch` manages a branch for a single `gitea_repository`.",
	}
}
