package gitea

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// Config is per-provider, specifies where to connect to gitea
type Config struct {
	Token      string
	Username   string
	Password   string
	BaseURL    string
	Insecure   bool
	CACertFile string
}

// Client returns a *gitea.Client to interact with the configured gitea instance
func (c *Config) Client() (interface{}, error) {

	if c.Token == "" && c.Username == "" {
		return nil, fmt.Errorf("either a token or a username needs to be used")
	}
	// Configure TLS/SSL
	var tlsConfig tls.Config
	// If a CACertFile has been specified, use that for cert validation
	if c.CACertFile != "" {
		caCert, err := ioutil.ReadFile(c.CACertFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// If configured as insecure, turn off SSL verification
	tlsConfig.InsecureSkipVerify = c.Insecure

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tlsConfig
	t.MaxIdleConnsPerHost = 100
	t.TLSHandshakeTimeout = 10 * time.Second

	httpClient := &http.Client{
		Transport: logging.NewTransport("Gitea", t),
	}

	if c.BaseURL == "" {
		c.BaseURL = "https://gitea.com"
	}

	var client *gitea.Client
	var err error
	if c.Token != "" {
		client, err = gitea.NewClient(c.BaseURL, gitea.SetToken(c.Token), gitea.SetHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}
	}

	if c.Username != "" {
		client, err = gitea.NewClient(c.BaseURL, gitea.SetBasicAuth(c.Username, c.Password), gitea.SetHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}
	}

	// Test the credentials by checking we can get information about the authenticated user.
	_, _, err = client.GetMyUserInfo()

	return client, err
}
