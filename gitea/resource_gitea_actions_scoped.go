package gitea

import (
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGiteaOrgActionsVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaOrgActionsVariableCreate,
		Read:   resourceGiteaOrgActionsVariableRead,
		Update: resourceGiteaOrgActionsVariableUpdate,
		Delete: resourceGiteaOrgActionsVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			actionOrgField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organisation owning the Actions variable.",
			},
			"variable_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Actions variable name.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Actions variable value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Actions variable description.",
			},
		},
		Description: "`gitea_org_actions_variable` manages an organisation-scoped Actions variable.\n\n" +
			"Import expects the resource ID in the form `org:variable_name`.",
	}
}

func resourceGiteaOrgActionsVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("variable_name").(string)
	_, err := client.CreateOrgActionVariable(org, name, gitea.CreateActionVariableOption{
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	d.SetId(buildTwoPartID(org, name))
	return resourceGiteaOrgActionsVariableRead(d, meta)
}

func resourceGiteaOrgActionsVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org, name, err := parseTwoPartID(d.Id(), actionOrgField, "variable_name")
	if err != nil {
		return err
	}
	variable, resp, err := client.GetOrgActionVariable(org, name)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set(actionOrgField, org)
	d.Set("variable_name", variable.Name)
	d.Set("value", variable.Data)
	d.Set(descriptionField, variable.Description)
	return nil
}

func resourceGiteaOrgActionsVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("variable_name").(string)
	_, err := client.UpdateOrgActionVariable(org, name, gitea.UpdateActionVariableOption{
		Name:        name,
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	return resourceGiteaOrgActionsVariableRead(d, meta)
}

func resourceGiteaOrgActionsVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org, name, err := parseTwoPartID(d.Id(), actionOrgField, "variable_name")
	if err != nil {
		return err
	}
	_, err = client.DeleteOrgActionVariable(org, name)
	return err
}

func resourceGiteaOrgActionsSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaOrgActionsSecretCreate,
		Read:   resourceGiteaOrgActionsSecretRead,
		Update: resourceGiteaOrgActionsSecretUpdate,
		Delete: resourceGiteaOrgActionsSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			actionOrgField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organisation owning the Actions secret.",
			},
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Actions secret name.",
			},
			"secret_value": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The Actions secret value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Actions secret description.",
			},
			createdAtField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Actions secret creation timestamp.",
			},
		},
		Description: "`gitea_org_actions_secret` manages an organisation-scoped Actions secret.\n\n" +
			"Import expects the resource ID in the form `org:secret_name`.\n" +
			"Because Gitea does not return secret values, `secret_value` must still be configured when importing.",
	}
}

func resourceGiteaOrgActionsSecretCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("secret_name").(string)
	_, err := client.CreateOrgActionSecret(org, name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	d.SetId(buildTwoPartID(org, name))
	return resourceGiteaOrgActionsSecretRead(d, meta)
}

func resourceGiteaOrgActionsSecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org, secretName, err := parseTwoPartID(d.Id(), actionOrgField, "secret_name")
	if err != nil {
		return err
	}
	secrets, err := collectPaginated(func(page int) ([]*gitea.Secret, error) {
		items, _, callErr := client.ListOrgActionSecret(org, gitea.ListOrgActionSecretOption{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 100},
		})
		return items, callErr
	})
	if err != nil {
		return err
	}
	var secret *gitea.Secret
	for _, item := range secrets {
		if item != nil && item.Name == secretName {
			secret = item
			break
		}
	}
	if secret == nil {
		d.SetId("")
		return nil
	}
	d.Set(actionOrgField, org)
	d.Set("secret_name", secret.Name)
	d.Set(descriptionField, secret.Description)
	d.Set(createdAtField, timeToString(secret.Created))
	return nil
}

func resourceGiteaOrgActionsSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("secret_name").(string)
	_, err := client.CreateOrgActionSecret(org, name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	return resourceGiteaOrgActionsSecretRead(d, meta)
}

func resourceGiteaOrgActionsSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org, name, err := parseTwoPartID(d.Id(), actionOrgField, "secret_name")
	if err != nil {
		return err
	}
	_, err = client.DeleteOrgActionSecret(org, name)
	return err
}

func resourceGiteaUserActionsVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaUserActionsVariableCreate,
		Read:   resourceGiteaUserActionsVariableRead,
		Update: resourceGiteaUserActionsVariableUpdate,
		Delete: resourceGiteaUserActionsVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"variable_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user-scoped Actions variable name.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user-scoped Actions variable value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user-scoped Actions variable description.",
			},
		},
		Description: "`gitea_user_actions_variable` manages a user-scoped Actions variable.\n\n" +
			"Import expects the resource ID in the form `variable_name`.",
	}
}

func resourceGiteaUserActionsVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	name := d.Get("variable_name").(string)
	_, err := client.CreateUserActionVariable(name, gitea.CreateActionVariableOption{
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceGiteaUserActionsVariableRead(d, meta)
}

func resourceGiteaUserActionsVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	variable, resp, err := client.GetUserActionVariable(d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set("variable_name", variable.Name)
	d.Set("value", variable.Data)
	d.Set(descriptionField, variable.Description)
	return nil
}

func resourceGiteaUserActionsVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	name := d.Get("variable_name").(string)
	_, err := client.UpdateUserActionVariable(name, gitea.UpdateActionVariableOption{
		Name:        name,
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	return resourceGiteaUserActionsVariableRead(d, meta)
}

func resourceGiteaUserActionsVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	_, err := client.DeleteUserActionVariable(d.Id())
	return err
}

func resourceGiteaUserActionsSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaUserActionsSecretCreate,
		Read:   resourceGiteaUserActionsSecretRead,
		Update: resourceGiteaUserActionsSecretUpdate,
		Delete: resourceGiteaUserActionsSecretDelete,
		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user-scoped Actions secret name.",
			},
			"secret_value": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The user-scoped Actions secret value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user-scoped Actions secret description.",
			},
		},
		Description: "`gitea_user_actions_secret` manages a user-scoped Actions secret.\n\n" +
			"This resource is write-only because the Gitea API does not expose a read/list endpoint for user-scoped Actions secrets.\n" +
			"Import is intentionally unsupported.",
	}
}

func resourceGiteaUserActionsSecretCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	name := d.Get("secret_name").(string)
	_, err := client.CreateUserActionSecret(name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceGiteaUserActionsSecretRead(d, meta)
}

func resourceGiteaUserActionsSecretRead(d *schema.ResourceData, meta interface{}) error {
	d.Set("secret_name", d.Get("secret_name").(string))
	d.Set(descriptionField, d.Get(descriptionField).(string))
	return nil
}

func resourceGiteaUserActionsSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	name := d.Get("secret_name").(string)
	_, err := client.CreateUserActionSecret(name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return err
	}
	return resourceGiteaUserActionsSecretRead(d, meta)
}

func resourceGiteaUserActionsSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	_, err := client.DeleteUserActionSecret(d.Id())
	return err
}
