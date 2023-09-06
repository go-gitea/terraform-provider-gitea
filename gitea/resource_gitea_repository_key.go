package gitea

import (
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	deployKeyRepoId   string = "repository"
	deployKeyName     string = "title"
	deployKeyKey      string = "key"
	deployKeyReadOnly string = "read_only"
)

func resourceRepoKeyIdParts(d *schema.ResourceData) (bool, int64, int64, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return false, 0, 0, nil
	}

	repoId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, 0, 0, err
	}
	keyId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, 0, 0, err
	}
	return true, repoId, keyId, err
}

func resourceRepoKeyRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	hasId, repoId, keyId, err := resourceRepoKeyIdParts(d)
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

	key, resp, err := client.GetDeployKey(repo.Owner.UserName, repo.Name, keyId)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	err = setRepoKeyResourceData(key, repoId, d)

	return
}

func resourceRepoKeyCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	repo, _, err := client.GetRepoByID(int64(d.Get(deployKeyRepoId).(int)))

	if err != nil {
		return err
	}

	dk, _, err := client.CreateDeployKey(repo.Owner.UserName, repo.Name, gitea.CreateKeyOption{
		Title:    d.Get(deployKeyName).(string),
		ReadOnly: d.Get(deployKeyReadOnly).(bool),
		Key:      d.Get(deployKeyKey).(string),
	})

	if err != nil {
		return err
	}

	setRepoKeyResourceData(dk, repo.ID, d)
	return nil
}

func respurceRepoKeyDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	hasId, repoId, keyId, err := resourceRepoKeyIdParts(d)
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

	client.DeleteDeployKey(repo.Owner.UserName, repo.Name, keyId)
	return nil
}

func setRepoKeyResourceData(dk *gitea.DeployKey, repoId int64, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d/%d", repoId, dk.ID))
	d.Set(deployKeyRepoId, repoId)
	d.Set(deployKeyReadOnly, dk.ReadOnly)
	d.Set(deployKeyKey, dk.Key)
	d.Set(deployKeyName, dk.Title)
	return
}

func resourceGiteaRepositoryKey() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepoKeyRead,
		Create: resourceRepoKeyCreate,
		Delete: respurceRepoKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			deployKeyRepoId: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the repository where the deploy key belongs to",
			},
			deployKeyKey: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Armored SSH key to add",
			},
			deployKeyReadOnly: {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Whether this key has read or read/write access",
			},
			deployKeyName: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the deploy key",
			},
		},
		Description: "`gitea_repository_key` manages a deploy key for a single gitea_repository.\n\n" +
			"Every key needs a unique name and unique key, i.e. no key can be added twice to the same repo",
	}
}
