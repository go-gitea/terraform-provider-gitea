package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	PublicKeyUser         string = "username"
	PublicKey             string = "key"
	PublicKeyReadOnlyFlag string = "read_only"
	PublicKeyTitle        string = "title"
	PublicKeyId           string = "id"
	PublicKeyFingerprint  string = "fingerprint"
	PublicKeyCreated      string = "created"
	PublicKeyType         string = "type"
)

func resourcePublicKeyRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response
	var pubKey *gitea.PublicKey

	pubKey, resp, err = client.GetPublicKey(id)

	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	err = setPublicKeyResourceData(pubKey, d)

	return
}

func resourcePublicKeyCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var pubKey *gitea.PublicKey

	opts := gitea.CreateKeyOption{
		Title:    d.Get(PublicKeyTitle).(string),
		Key:      d.Get(PublicKey).(string),
		ReadOnly: d.Get(PublicKeyReadOnlyFlag).(bool),
	}

	pubKey, _, err = client.AdminCreateUserPublicKey(d.Get(PublicKeyUser).(string), opts)

	err = setPublicKeyResourceData(pubKey, d)

	return
}

func resourcePublicKeyUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	// update = recreate
	resourcePublicKeyDelete(d, meta)
	resourcePublicKeyCreate(d, meta)
	return
}

func resourcePublicKeyDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response

	resp, err = client.AdminDeleteUserPublicKey(d.Get(PublicKeyUser).(string), int(id))

	if err != nil {
		if resp.StatusCode == 404 {
			return
		} else {
			return err
		}
	}

	return
}

func setPublicKeyResourceData(pubKey *gitea.PublicKey, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d", pubKey.ID))
	d.Set(PublicKeyUser, d.Get(PublicKeyUser).(string))
	d.Set(PublicKey, d.Get(PublicKey).(string))
	d.Set(PublicKeyTitle, pubKey.Title)
	d.Set(PublicKeyReadOnlyFlag, d.Get(PublicKeyReadOnlyFlag).(bool))
	d.Set(PublicKeyCreated, pubKey.Created)
	d.Set(PublicKeyFingerprint, pubKey.Fingerprint)
	d.Set(PublicKeyType, pubKey.KeyType)
	return
}

func resourceGiteaPublicKey() *schema.Resource {
	return &schema.Resource{
		Read:   resourcePublicKeyRead,
		Create: resourcePublicKeyCreate,
		Update: resourcePublicKeyUpdate,
		Delete: resourcePublicKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Title of the key to add",
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "An armored SSH key to add",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     false,
				Description: "Describe if the key has only read access or read/write",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				ForceNew:    true,
				Description: "User to associate with the added key",
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Description: "`gitea_public_key` manages ssh key that are associated with users.",
	}
}
