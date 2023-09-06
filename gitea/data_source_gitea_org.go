package gitea

import (
	"fmt"
	"log"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaOrg() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGiteaOrgRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	var org *gitea.Organization
	var err error

	log.Printf("[INFO] Reading Gitea Org")

	nameData, nameOk := d.GetOk("name")

	if !nameOk {
		return fmt.Errorf("name of org must be passed")
	}
	name := strings.ToLower(nameData.(string))

	org, _, err = client.GetOrg(name)
	if err != nil {
		return err
	}

	d.Set("id", org.ID)
	d.Set("name", org.UserName)
	d.Set("full_name", org.FullName)
	d.Set("avatar_url", org.AvatarURL)
	d.Set("location", org.Location)
	d.Set("website", org.Website)
	d.Set("description", org.Description)
	d.Set("visibility", org.Visibility)

	d.SetId(fmt.Sprintf("%d", org.ID))

	return nil
}
