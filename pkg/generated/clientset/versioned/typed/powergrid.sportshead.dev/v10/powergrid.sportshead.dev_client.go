// Code generated by client-gen. DO NOT EDIT.

package v10

import (
	"net/http"

	v10 "github.com/sportshead/powergrid/pkg/apis/powergrid.sportshead.dev/v10"
	"github.com/sportshead/powergrid/pkg/generated/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type PowergridV10Interface interface {
	RESTClient() rest.Interface
	CommandsGetter
}

// PowergridV10Client is used to interact with features provided by the powergrid.sportshead.dev group.
type PowergridV10Client struct {
	restClient rest.Interface
}

func (c *PowergridV10Client) Commands(namespace string) CommandInterface {
	return newCommands(c, namespace)
}

// NewForConfig creates a new PowergridV10Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*PowergridV10Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new PowergridV10Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*PowergridV10Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &PowergridV10Client{client}, nil
}

// NewForConfigOrDie creates a new PowergridV10Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *PowergridV10Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new PowergridV10Client for the given RESTClient.
func New(c rest.Interface) *PowergridV10Client {
	return &PowergridV10Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v10.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *PowergridV10Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}