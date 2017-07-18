package clients

import (
	"encoding/json"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/api/types"
	"github.com/stitchfix/flotilla-os/config"
	"net/url"
)

type RegistryClient interface {
	ListRepositories() ([]string, error)
	ListTags(image string) ([]string, error)
}

type registryClient struct {
	registry dockerRegistry
	repo     dockerRepository
}

type dockerRegistry interface {
	Repositories(ctx context.Context, entries []string, last string) (int, error)
}

type dockerRepository interface {
	All(ctx context.Context) ([]string, error)
}

type simpleCredentialStore struct {
	auth types.AuthConfig
}

func (scs simpleCredentialStore) Basic(u *url.URL) (string, string) {
	return scs.auth.Username, scs.auth.Password
}

func (scs simpleCredentialStore) RefreshToken(u *url.URL, service string) string {
	return scs.auth.IdentityToken
}

func (scs simpleCredentialStore) SetRefreshToken(*url.URL, string, string) {
}

func loadAuthConfigs(c config.Config) map[string]types.AuthConfig {

	if !c.IsSet("docker_auth") {
		// read configs from string

	}
}

func NewRegistryClient(c config.Config) (RegistryClient, error) {
	var rc registryClient
	var baseURL string
	if !c.IsSet("registry.base_url") {
		baseURL = "https://index.docker.io"
	} else {
		baseURL = c.GetString("registry.base_url")
	}

	// Need a new transport object
	cm := challenge.NewSimpleManager()
	tokenHandler := auth.NewTokenHandler(nil, nil, "", "pull", "push")
	authorizer := auth.NewAuthorizer(cm, tokenHandler)

	tsp := transport.NewTransport(nil, authorizer)

	r, err := client.NewRegistry(context.Background(), baseURL, tsp)
	rc.registry = r
	return &rc, err
}

func (rc *registryClient) ListRepositories() ([]string, error) {
	repos := make([]string, 8)
	_, err := rc.registry.Repositories(context.Background(), repos, "")
	return repos, err
}

func (rc *registryClient) ListTags(image string) ([]string, error) {
	var tags []string
	return tags, nil
}
