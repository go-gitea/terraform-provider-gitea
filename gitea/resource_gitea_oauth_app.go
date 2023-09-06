package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	oauth2KeyName               string = "name"
	oauth2KeyConfidentialClient string = "confidential_client"
	oauth2KeyRedirectURIs       string = "redirect_uris"
	oauth2KeyClientId           string = "client_id"
	oauth2KeyClientSecret       string = "client_secret"
)

func resourceGiteaOauthApp() *schema.Resource {
	return &schema.Resource{
		Read:   resourceOauth2AppRead,
		Create: resourceOauth2AppUpcreate,
		Update: resourceOauth2AppUpcreate,
		Delete: resourceOauth2AppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			oauth2KeyName: {
				Required:    true,
				Type:        schema.TypeString,
				Description: "OAuth Application name",
			},
			oauth2KeyRedirectURIs: {
				Required: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Accepted redirect URIs",
			},
			oauth2KeyConfidentialClient: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to false, it will be a public client (PKCE will be required)",
			},
			oauth2KeyClientId: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "OAuth2 Application client id",
			},
			oauth2KeyClientSecret: {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Oauth2 Application client secret",
			},
		},
		Description: "Handling [gitea oauth application](https://docs.gitea.io/en-us/oauth2-provider/) resources",
	}
}

func ExpandStringList(configured []interface{}) []string {
	res := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			res = append(res, (v.(string)))
		}
	}
	return res
}

func CollapseStringList(strlist []string) []interface{} {
	res := make([]interface{}, 0, len(strlist))
	for _, v := range strlist {
		res = append(res, v)
	}
	return res
}

func resourceOauth2AppUpcreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	redirectURIsSchema, redirectURIsSchemaOk := d.Get(oauth2KeyRedirectURIs).(*schema.Set)

	if !redirectURIsSchemaOk {
		return fmt.Errorf("attribute %s must be set to a set of strings", oauth2KeyRedirectURIs)
	}

	redirectURIs := ExpandStringList(redirectURIsSchema.List())

	name, nameOk := d.Get(oauth2KeyName).(string)

	if !nameOk {
		return fmt.Errorf("attribute %s must be set and must be a string", oauth2KeyName)
	}

	confidentialClient, confidentialClientOk := d.Get(oauth2KeyConfidentialClient).(bool)

	if !confidentialClientOk {
		return fmt.Errorf("attribute %s must be set and must be a bool", oauth2KeyConfidentialClient)
	}

	opts := gitea.CreateOauth2Option{
		Name:               name,
		ConfidentialClient: confidentialClient,
		RedirectURIs:       redirectURIs,
	}

	var oauth2 *gitea.Oauth2

	if d.IsNewResource() {
		oauth2, _, err = client.CreateOauth2(opts)
	} else {
		oauth2, err = searchOauth2AppByClientId(client, d.Id())

		if err != nil {
			return err
		}

		oauth2, _, err = client.UpdateOauth2(oauth2.ID, opts)
	}

	if err != nil {
		return
	}

	err = setOAuth2ResourceData(oauth2, d)

	return
}

func searchOauth2AppByClientId(c *gitea.Client, id string) (res *gitea.Oauth2, err error) {
	page := 1

	for {
		apps, _, err := c.ListOauth2(gitea.ListOauth2Option{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 50,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(apps) == 0 {
			return nil, fmt.Errorf("no oauth client can be found by id '%s'", id)
		}

		for _, app := range apps {
			if app.ClientID == id {
				return app, nil
			}
		}

		page += 1
	}
}

func resourceOauth2AppRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	app, err := searchOauth2AppByClientId(client, d.Id())

	if err != nil {
		return err
	}

	err = setOAuth2ResourceData(app, d)

	return
}

func resourceOauth2AppDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	app, err := searchOauth2AppByClientId(client, d.Id())

	if err != nil {
		return err
	}

	_, err = client.DeleteOauth2(app.ID)

	return
}

func setOAuth2ResourceData(app *gitea.Oauth2, d *schema.ResourceData) (err error) {
	d.SetId(app.ClientID)

	for k, v := range map[string]interface{}{
		oauth2KeyName:               app.Name,
		oauth2KeyConfidentialClient: app.ConfidentialClient,
		oauth2KeyRedirectURIs:       schema.NewSet(schema.HashString, CollapseStringList(app.RedirectURIs)),
		oauth2KeyClientId:           app.ClientID,
	} {
		err = d.Set(k, v)
		if err != nil {
			return
		}
	}

	if app.ClientSecret != "" {
		// Gitea API only reports client secrets if the resource is newly created
		d.Set(oauth2KeyClientSecret, app.ClientSecret)
	}

	return
}
