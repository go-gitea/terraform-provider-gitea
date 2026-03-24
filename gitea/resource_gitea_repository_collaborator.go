package gitea

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	collabOwner      string = "owner"
	collabRepo       string = "repo"
	collabUsername   string = "username"
	collabPermission string = "permission"
)

func resourceRepositoryCollaboratorCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	owner := d.Get(collabOwner).(string)
	repo := d.Get(collabRepo).(string)
	username := d.Get(collabUsername).(string)
	permission := gitea.AccessMode(d.Get(collabPermission).(string))

	_, err = client.AddCollaborator(owner, repo, username, gitea.AddCollaboratorOption{
		Permission: &permission,
	})
	if err != nil {
		return
	}

	err = setRepositoryCollaboratorData(d, owner, repo, username, string(permission))

	return
}

func resourceRepositoryCollaboratorRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	owner, repo, username, parseErr := parseCollaboratorID(d.Id())
	if parseErr != nil {
		return parseErr
	}

	isCollab, resp, err := client.IsCollaborator(owner, repo, username)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}
	if !isCollab {
		d.SetId("")
		return nil
	}

	permResult, _, err := client.CollaboratorPermission(owner, repo, username)
	if err != nil {
		return err
	}

	err = setRepositoryCollaboratorData(d, owner, repo, username, string(permResult.Permission))

	return
}

func resourceRepositoryCollaboratorUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	owner := d.Get(collabOwner).(string)
	repo := d.Get(collabRepo).(string)
	username := d.Get(collabUsername).(string)
	permission := gitea.AccessMode(d.Get(collabPermission).(string))

	_, err = client.AddCollaborator(owner, repo, username, gitea.AddCollaboratorOption{
		Permission: &permission,
	})
	if err != nil {
		return
	}

	err = setRepositoryCollaboratorData(d, owner, repo, username, string(permission))

	return
}

func resourceRepositoryCollaboratorDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	owner := d.Get(collabOwner).(string)
	repo := d.Get(collabRepo).(string)
	username := d.Get(collabUsername).(string)

	_, err = client.DeleteCollaborator(owner, repo, username)

	return
}

func parseCollaboratorID(id string) (owner, repo, username string, err error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("invalid collaborator ID format: %s (expected owner/repo/username)", id)
		return
	}
	owner = parts[0]
	repo = parts[1]
	username = parts[2]
	return
}

func setRepositoryCollaboratorData(d *schema.ResourceData, owner, repo, username, permission string) (err error) {
	d.SetId(fmt.Sprintf("%s/%s/%s", owner, repo, username))
	d.Set(collabOwner, owner)
	d.Set(collabRepo, repo)
	d.Set(collabUsername, username)
	d.Set(collabPermission, permission)
	return
}

func resourceGiteaRepositoryCollaborator() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepositoryCollaboratorRead,
		Create: resourceRepositoryCollaboratorCreate,
		Update: resourceRepositoryCollaboratorUpdate,
		Delete: resourceRepositoryCollaboratorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The owner of the repository (user or organization).",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the collaborator to add.",
			},
			"permission": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "read",
				Description: "The permission level for the collaborator: `read`, `write`, or `admin`.",
			},
		},
		Description: "`gitea_repository_collaborator` manages a single collaborator's access to a repository without requiring team creation.",
	}
}
