package gitea

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGiteaRepositoryActionsSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaRepositoryActionsSecretCreate,
		Read:   resourceGiteaRepositoryActionsSecretRead,
		Update: resourceGiteaRepositoryActionsSecretUpdate,
		Delete: resourceGiteaRepositoryActionsSecretDelete,
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
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the secret.",
			},
			"secret_value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of the secret.",
				Sensitive:   true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'actions_secret' creation.",
			},
		},
	}
}

func resourceGiteaRepositoryActionsSecretCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return fmt.Errorf("name of repo owner must be passed")
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	repositoryData, nameOk := d.GetOk("repository")
	if !nameOk {
		return fmt.Errorf("CREATE name of repo must be passed")
	}
	repository := strings.ToLower(repositoryData.(string))

	secretNameData, nameOk := d.GetOk("secret_name")
	if !nameOk {
		return fmt.Errorf("secret_name must be passed")
	}
	secretName := secretNameData.(string)

	valueData, nameOk := d.GetOk("secret_value")
	if !nameOk {
		return fmt.Errorf("value must be passed")
	}
	value := valueData.(string)

	_, err := client.CreateRepoActionSecret(repoOwner, repository, gitea.CreateSecretOption{
		Name: secretName,
		Data: value,
	})
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(repoOwner, repository, secretName))

	return resourceGiteaRepositoryActionsSecretRead(d, meta)
}

func resourceGiteaRepositoryActionsSecretUpdate(d *schema.ResourceData, meta interface{}) error {
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

	variableNameData, nameOk := d.GetOk("secret_name")
	if !nameOk {
		return fmt.Errorf("secret_name of repo must be passed")
	}
	variableName := variableNameData.(string)

	valueData, nameOk := d.GetOk("secret_value")
	if !nameOk {
		return fmt.Errorf("secret_value must be passed")
	}
	value := valueData.(string)

	_, err := client.CreateRepoActionSecret(repoOwner, repository, gitea.CreateSecretOption{
		Name: variableName,
		Data: value,
	})
	if err != nil {
		return err
	}

	return resourceGiteaRepositoryActionsSecretRead(d, meta)
}

func resourceGiteaRepositoryActionsSecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwner, repository, secretName, _ := parseThreePartID(d.Id(), "repository_owner", "repository", "secret_name")

	var requestedSecret *gitea.Secret

	page := 0
	for requestedSecret == nil {
		page = page + 1

		secrets, _, _ := client.ListRepoActionSecret(repoOwner, repository, gitea.ListRepoActionSecretOption{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 100,
			},
		})

		if len(secrets) == 0 {
			d.SetId("")
			return nil
		}

		for _, secret := range secrets {
			if secret.Name == secretName {
				requestedSecret = secret
				break
			}
		}
	}

	createdAtData, dateOk := d.GetOk("created_at")

	if requestedSecret == nil {
		d.SetId("")
		return nil
	}

	if dateOk {
		if requestedSecret.Created.String() != createdAtData.(string) {
			d.SetId("")
			return nil
		}
	}

	if err := d.Set("repository_owner", repoOwner); err != nil {
		return err
	}

	if err := d.Set("repository", repository); err != nil {
		return err
	}

	if err := d.Set("secret_name", secretName); err != nil {
		return err
	}

	if err := d.Set("created_at", requestedSecret.Created.String()); err != nil {
		return err
	}

	return nil
}

func resourceGiteaRepositoryActionsSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	repoOwner, repository, secretName, _ := parseThreePartID(d.Id(), "repository_owner", "repository", "secret_name")

	_, err := client.DeleteRepoActionSecret(repoOwner, repository, secretName)

	return err
}
