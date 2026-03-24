package gitea

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	GPGKeyId      string = "id"
	GPGKeyArmored string = "armored_public_key"
	GPGKeyGPGId   string = "key_id"
)

func resourceGPGKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response
	var pubKey *gitea.GPGKey

	pubKey, resp, err = client.GetGPGKey(id)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	err = setGPGKeyResourceData(pubKey, d)

	return diag.FromErr(err)
}

func resourceGPGKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	var pubKey *gitea.GPGKey
	var err error

	opts := gitea.CreateGPGKeyOption{
		ArmoredKey: d.Get(GPGKeyArmored).(string),
	}

	pubKey, _, err = client.CreateGPGKey(opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating gpg key: %w", err))
	}

	err = setGPGKeyResourceData(pubKey, d)

	return diag.FromErr(err)
}

func resourceGPGKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting gpg key: invalid id: %w", err))
	}

	var resp *gitea.Response

	resp, err = client.DeleteGPGKey(id)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil
		} else {
			return diag.FromErr(fmt.Errorf("error deleting gpg key: %w", err))
		}
	}

	return nil
}

func setGPGKeyResourceData(pubKey *gitea.GPGKey, d *schema.ResourceData) error {
	d.SetId(fmt.Sprintf("%d", pubKey.ID))
	d.Set(GPGKeyArmored, pubKey.PublicKey)
	d.Set(GPGKeyGPGId, pubKey.KeyID)

	return nil
}

func resourceGiteaGPGKey() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceGPGKeyRead,
		CreateContext: resourceGPGKeyCreate,
		DeleteContext: resourceGPGKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"armored_public_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "An armored GPG public key",
			},
			"key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the GPG key",
			},
		},
		Description: "`gitea_gpg_key` manages gpg keys that are associated with the current user.",
	}
}
