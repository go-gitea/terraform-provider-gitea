package gitea

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGiteaRepositoryActionsVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaRepositoryActionsVariableCreate,
		Read:   resourceGiteaRepositoryActionsVariableRead,
		Update: resourceGiteaRepositoryActionsVariableUpdate,
		Delete: resourceGiteaRepositoryActionsVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository_owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Owner of the repository.",
			},
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the repository.",
			},
			"variable_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the variable.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of the variable.",
			},
		},
	}
}

func resourceGiteaRepositoryActionsVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return fmt.Errorf("name of repo owner must be passed")
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	nameData, nameOk := d.GetOk("repository")
	if !nameOk {
		return fmt.Errorf("CREATE name of repo must be passed")
	}
	name := strings.ToLower(nameData.(string))

	variableNameData, nameOk := d.GetOk("variable_name")
	if !nameOk {
		return fmt.Errorf("variable_name of repo must be passed")
	}
	variableName := variableNameData.(string)

	valueData, nameOk := d.GetOk("value")
	if !nameOk {
		return fmt.Errorf("value must be passed")
	}
	value := valueData.(string)

	_, err := client.CreateRepoActionVariable(repoOwner, name, variableName, value)
	if err != nil {
		return err
	}
	d.SetId(buildThreePartID(repoOwner, name, variableName))

	return resourceGiteaRepositoryActionsVariableRead(d, meta)
}

func resourceGiteaRepositoryActionsVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return fmt.Errorf("name of repo owner must be passed")
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	repositoryData, nameOk := d.GetOk("repository")
	if !nameOk {
		return fmt.Errorf("READ name of repo must be passed")
	}
	repository := strings.ToLower(repositoryData.(string))

	variableNameData, nameOk := d.GetOk("variable_name")
	if !nameOk {
		return fmt.Errorf("READ variable_name of repo must be passed")
	}
	variableName := variableNameData.(string)

	valueData, nameOk := d.GetOk("value")
	if !nameOk {
		return fmt.Errorf("value must be passed")
	}
	value := valueData.(string)

	_, err := client.UpdateRepoActionVariable(repoOwner, repository, variableName, value)
	if err != nil {
		return err
	}

	return resourceGiteaRepositoryActionsVariableRead(d, meta)
}

func resourceGiteaRepositoryActionsVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwner, repository, variableName, err := parseThreePartID(d.Id(), "repository_owner", "repository", "variable_name")
	if err != nil {
		return err
	}

	variable, resp, err := client.GetRepoActionVariable(repoOwner, repository, variableName)

	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	if err = d.Set("repository_owner", repoOwner); err != nil {
		return err
	}

	if err = d.Set("repository", repository); err != nil {
		return err
	}

	if err = d.Set("variable_name", variableName); err != nil {
		return err
	}

	if err = d.Set("value", variable.Value); err != nil {
		return err
	}

	return nil
}

func resourceGiteaRepositoryActionsVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return fmt.Errorf("name of repo owner must be passed")
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	repositoryData, nameOk := d.GetOk("repository")
	if !nameOk {
		return fmt.Errorf("name of repo must be passed")
	}
	repository := strings.ToLower(repositoryData.(string))

	variableNameData, nameOk := d.GetOk("variable_name")
	if !nameOk {
		return fmt.Errorf("variable_name must be passed")
	}
	variableName := strings.ToLower(variableNameData.(string))

	_, err := client.DeleteRepoActionVariable(repoOwner, repository, variableName)

	return err
}

// format the strings into an id `a:b:c`
func buildThreePartID(a, b, c string) string {
	return fmt.Sprintf("%s:%s:%s", a, b, c)
}
func parseThreePartID(id, left, center, right string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected ID format (%q). Expected %s:%s:%s", id, left, center, right)
	}

	return parts[0], parts[1], parts[2], nil
}
