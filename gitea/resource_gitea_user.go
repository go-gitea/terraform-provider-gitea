package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	userName                string = "username"
	userLoginName           string = "login_name"
	userEmail               string = "email"
	userFullName            string = "full_name"
	userPassword            string = "password"
	userMustChangePassword  string = "must_change_password"
	userSendNotification    string = "send_notification"
	userVisibility          string = "visibility"
	userDescription         string = "description"
	userLocation            string = "location"
	userActive              string = "active"
	userAdmin               string = "admin"
	userAllowGitHook        string = "allow_git_hook"
	userAllowLocalImport    string = "allow_import_local"
	userMaxRepoCreation     string = "max_repo_creation"
	userPhorbitLogin        string = "prohibit_login"
	userAllowCreateOrgs     string = "allow_create_organization"
	userRestricted          string = "restricted"
	userForcePasswordChange string = "force_password_change"
)

func resourceUserRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	var resp *gitea.Response
	var user *gitea.User

	user, resp, err = client.GetUserByID(id)

	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	err = setUserResourceData(user, d)

	return
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var user *gitea.User
	visibility := gitea.VisibleType(d.Get(userVisibility).(string))
	changePassword := d.Get(userMustChangePassword).(bool)

	opts := gitea.CreateUserOption{
		SourceID:           0,
		LoginName:          d.Get(userLoginName).(string),
		Username:           d.Get(userName).(string),
		FullName:           d.Get(userFullName).(string),
		Email:              d.Get(userEmail).(string),
		Password:           d.Get(userPassword).(string),
		MustChangePassword: &changePassword,
		SendNotify:         d.Get(userSendNotification).(bool),
		Visibility:         &visibility,
	}

	user, _, err = client.AdminCreateUser(opts)
	if err != nil {
		return
	}

	d.SetId(fmt.Sprintf("%d", user.ID))

	err = resourceUserUpdate(d, meta)

	return
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	var resp *gitea.Response
	var user *gitea.User

	user, resp, err = client.GetUserByID(id)

	if err != nil {
		if resp.StatusCode == 404 {
			resourceUserCreate(d, meta)
		} else {
			return err
		}
	}

	mail := d.Get(userEmail).(string)
	fullName := d.Get(userFullName).(string)
	description := d.Get(userDescription).(string)
	changePassword := d.Get(userMustChangePassword).(bool)
	location := d.Get(userLocation).(string)
	active := d.Get(userActive).(bool)
	admin := d.Get(userAdmin).(bool)
	allowHook := d.Get(userAllowGitHook).(bool)
	allowImport := d.Get(userAllowLocalImport).(bool)
	maxRepoCreation := d.Get(userMaxRepoCreation).(int)
	accessDenied := d.Get(userPhorbitLogin).(bool)
	allowOrgs := d.Get(userAllowCreateOrgs).(bool)
	restricted := d.Get(userRestricted).(bool)
	visibility := gitea.VisibleType(d.Get(userVisibility).(string))

	if d.Get(userForcePasswordChange).(bool) {
		opts := gitea.EditUserOption{
			SourceID:                0,
			LoginName:               d.Get(userLoginName).(string),
			Email:                   &mail,
			FullName:                &fullName,
			Password:                d.Get(userPassword).(string),
			Description:             &description,
			MustChangePassword:      &changePassword,
			Location:                &location,
			Active:                  &active,
			Admin:                   &admin,
			AllowGitHook:            &allowHook,
			AllowImportLocal:        &allowImport,
			MaxRepoCreation:         &maxRepoCreation,
			ProhibitLogin:           &accessDenied,
			AllowCreateOrganization: &allowOrgs,
			Restricted:              &restricted,
			Visibility:              &visibility,
		}
		_, err = client.AdminEditUser(d.Get(userName).(string), opts)

		if err != nil {
			return err
		}

	} else {
		opts := gitea.EditUserOption{
			SourceID:                0,
			LoginName:               d.Get(userLoginName).(string),
			Email:                   &mail,
			FullName:                &fullName,
			Description:             &description,
			MustChangePassword:      &changePassword,
			Location:                &location,
			Active:                  &active,
			Admin:                   &admin,
			AllowGitHook:            &allowHook,
			AllowImportLocal:        &allowImport,
			MaxRepoCreation:         &maxRepoCreation,
			ProhibitLogin:           &accessDenied,
			AllowCreateOrganization: &allowOrgs,
			Restricted:              &restricted,
			Visibility:              &visibility,
		}
		_, err = client.AdminEditUser(d.Get(userName).(string), opts)

		if err != nil {
			return err
		}
	}

	user, _, err = client.GetUserByID(id)

	err = setUserResourceData(user, d)

	return
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	var resp *gitea.Response

	resp, err = client.AdminDeleteUser(d.Get(userName).(string))

	if err != nil {
		if resp.StatusCode == 404 {
			return
		} else {
			return err
		}
	}

	return
}

func setUserResourceData(user *gitea.User, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d", user.ID))
	d.Set(userName, user.UserName)
	d.Set(userEmail, user.Email)
	d.Set(userFullName, user.FullName)
	d.Set(userAdmin, user.IsAdmin)
	d.Set("created", user.Created)
	d.Set("avatar_url", user.AvatarURL)
	d.Set("last_login", user.LastLogin)
	d.Set("language", user.Language)
	d.Set(userLoginName, d.Get(userLoginName).(string))
	d.Set(userMustChangePassword, d.Get(userMustChangePassword).(bool))
	d.Set(userSendNotification, d.Get(userSendNotification).(bool))
	d.Set(userVisibility, d.Get(userVisibility).(string))
	d.Set(userDescription, d.Get(userDescription).(string))
	d.Set(userLocation, d.Get(userLocation).(string))
	d.Set(userActive, d.Get(userActive).(bool))
	d.Set(userAllowGitHook, d.Get(userAllowGitHook).(bool))
	d.Set(userAllowLocalImport, d.Get(userAllowLocalImport).(bool))
	d.Set(userMaxRepoCreation, d.Get(userMaxRepoCreation).(int))
	d.Set(userPhorbitLogin, d.Get(userPhorbitLogin).(bool))
	d.Set(userAllowCreateOrgs, d.Get(userAllowCreateOrgs).(bool))
	d.Set(userRestricted, d.Get(userRestricted).(bool))
	d.Set(userForcePasswordChange, d.Get(userForcePasswordChange).(bool))

	return
}

func resourceGiteaUser() *schema.Resource {
	return &schema.Resource{
		Read:   resourceUserRead,
		Create: resourceUserCreate,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Username of the user to be created",
			},
			"login_name": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "The login name can differ from the username",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "E-Mail Address of the user",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Required:    false,
				Description: "Full name of the user",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Sensitive:   true,
				Description: "Password to be set for the user",
			},
			"must_change_password": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     true,
				Description: "Flag if the user should change the password after first login",
			},
			"send_notification": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     true,
				Description: "Flag to send a notification about the user creation to the defined `email`",
			},
			"visibility": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				Default:     "public",
				Description: "Visibility of the user. Can be `public`, `limited` or `private`",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				Default:     "",
				Description: "A description of the user",
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				Required: false,
				Default:  "",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     true,
				Description: "Flag if this user should be active or not",
			},
			"admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     false,
				Description: "Flag if this user should be an administrator or not",
			},
			"allow_git_hook": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  true,
			},
			"allow_import_local": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  true,
			},
			"max_repo_creation": {
				Type:     schema.TypeInt,
				Optional: true,
				Required: false,
				Default:  -1,
			},
			"prohibit_login": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     false,
				Description: "Flag if the user should not be allowed to log in (bot user)",
			},
			"allow_create_organization": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  true,
			},
			"restricted": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  false,
			},
			"force_password_change": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     false,
				Description: "Flag if the user defined password should be overwritten or not",
			},
		},
		Description: "`gitea_user` manages a native gitea user.\n\n" +
			"If you are using OIDC or other kinds of authentication mechanisms you can still try to manage" +
			"ssh keys or other ressources this way",
	}
}
