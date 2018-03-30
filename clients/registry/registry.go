package registry

import (
	"context"
	"fmt"
	dockercfg "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	dist "github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/pkg/homedir"
	"github.com/docker/go-connections/sockets"
	"github.com/moby/moby/registry"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

//
// Client has the sole purpose of validating that an image exists
// and that the use can access it. This -dramatically- reduces the debug
// cycle, particularly with ill-formed or inaccessible image names
//
type Client interface {
	IsImageValid(imageRef string) (bool, error)
}

type registryClient struct {
	tagFetcher ImageTagFetcher
}

// ImageTagFetcher low-level tag fetcher interface
type ImageTagFetcher interface {
	TagsForImage(imageRef reference.Reference) ([]string, error)
}

type imageTagFetcher struct {
	registryService registry.Service
	authConfigs     map[string]types.AuthConfig
}

func loadAuthConfigs(c config.Config) (map[string]types.AuthConfig, error) {
	// Use docker's cli to load auth configs from file or string
	var err error
	var cfile *configfile.ConfigFile
	if !c.IsSet("docker_auth") {
		// read configs from string
		configDir := filepath.Join(homedir.Get(), ".docker")
		cfile, err = dockercfg.Load(configDir)
		if err != nil {
			return nil, err
		}
	} else {
		cfile, err = dockercfg.LoadFromReader(
			strings.NewReader(c.GetString("docker_auth")))
		if err != nil {
			return nil, err
		}
	}
	return cfile.GetAuthConfigs(), nil
}

//
// NewRegistryClient returns a new RegistryClient that knows how to validate image references
//
func NewRegistryClient(c config.Config) (Client, error) {
	var (
		err error
		rc  registryClient
		tf  imageTagFetcher
	)

	tf.registryService = registry.NewService(registry.ServiceOptions{})
	tf.authConfigs, err = loadAuthConfigs(c)
	if err != nil {
		return &rc, errors.Wrap(err, "error loading docker authentication configuration")
	}
	rc.tagFetcher = &tf
	return &rc, nil
}

func (rc *registryClient) IsImageValid(imageRef string) (bool, error) {
	ref, err := reference.ParseAnyReference(imageRef)
	if err != nil {
		return false, errors.Wrapf(err, "issue parsing image reference [%s]", imageRef)
	}
	taggedRef, ok := ref.(reference.Tagged)
	if !ok {
		return false, errors.Errorf("can't get tag from image reference [%s]", imageRef)
	}

	tags, err := rc.tagFetcher.TagsForImage(ref)
	if err != nil {
		return false, errors.Wrapf(err, "issue fetching tags for image reference [%s]", imageRef)
	}

	tag := taggedRef.Tag()
	for _, t := range tags {
		if tag == t {
			return true, nil
		}
	}
	return false, nil
}

func (tf *imageTagFetcher) TagsForImage(imageRef reference.Reference) ([]string, error) {
	var (
		err  error
		tags []string
	)

	ctx := context.Background()
	repo, err := tf.repositoryForRef(ctx, imageRef)
	if err != nil {
		return tags, errors.Wrap(err, "issue resolving repository")
	}

	tags, err = repo.Tags(ctx).All(ctx)
	if err != nil {
		return tags, errors.Wrap(err, "issue fetching tags from image repository")
	}
	return tags, nil
}

//
// Pulled whole or in part from github.com/moby/moby/distribution/registry.go
//
func (tf *imageTagFetcher) repositoryForRef(ctx context.Context, imageRef reference.Reference) (dist.Repository, error) {
	// Parse the imageName
	namedRef, _ := imageRef.(reference.Named)

	// Resolve repository
	repoInfo, err := tf.registryService.ResolveRepository(namedRef)
	if err != nil {
		return nil, errors.Wrap(err, "issue resolving repository")
	}

	// Get auth for repo
	authConfig := registry.ResolveAuthConfig(tf.authConfigs, repoInfo.Index)

	// Get endpoints
	endpoints, err := tf.registryService.LookupPullEndpoints(
		reference.Domain(repoInfo.Name))
	if err != nil {
		return nil, errors.Wrapf(err, "issue looking up pull endpoints for [%s]", repoInfo.Name.String())
	}

	// Get v2 repo - only v2 is supported
	var (
		repository dist.Repository
		lastError  error
	)

	for _, endpoint := range endpoints {
		if endpoint.Version == registry.APIVersion1 {
			continue
		}

		repository, lastError = tf.newV2Repository(
			ctx, repoInfo, endpoint, nil, &authConfig, "pull")
		if lastError == nil {
			break
		}
	}
	if lastError != nil {
		return repository, errors.WithStack(err)
	}
	return repository, nil
}

func (tf *imageTagFetcher) newV2Repository(
	ctx context.Context,
	repoInfo *registry.RepositoryInfo,
	endpoint registry.APIEndpoint,
	metaHeaders http.Header,
	authConfig *types.AuthConfig, actions ...string) (dist.Repository, error) {

	repoName := repoInfo.Name.Name()
	if endpoint.TrimHostname {
		repoName = reference.Path(repoInfo.Name)
	}

	direct := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	base := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                direct.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     endpoint.TLSConfig,
		DisableKeepAlives:   true,
	}

	proxyDialer, err := sockets.DialerFromEnvironment(direct)
	if err == nil {
		base.Dial = proxyDialer.Dial
	}

	modifiers := registry.DockerHeaders(dockerversion.DockerUserAgent(ctx), metaHeaders)
	authTransport := transport.NewTransport(base, modifiers...)

	challengeManager, _, err := registry.PingV2Registry(endpoint.URL, authTransport)
	if err != nil {
		return nil, errors.Wrapf(err, "issue pinging v2 registry with url [%v]", endpoint.URL)
	}

	if authConfig.RegistryToken != "" {
		passThruTokenHandler := &existingTokenHandler{token: authConfig.RegistryToken}
		modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, passThruTokenHandler))
	} else {
		scope := auth.RepositoryScope{
			Repository: repoName,
			Actions:    actions,
			Class:      repoInfo.Class,
		}

		creds := registry.NewStaticCredentialStore(authConfig)
		tokenHandlerOptions := auth.TokenHandlerOptions{
			Transport:   authTransport,
			Credentials: creds,
			Scopes:      []auth.Scope{scope},
			ClientID:    registry.AuthClientID,
		}
		tokenHandler := auth.NewTokenHandlerWithOptions(tokenHandlerOptions)
		basicHandler := auth.NewBasicHandler(creds)
		modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler))
	}
	tr := transport.NewTransport(base, modifiers...)

	repoNameRef, err := reference.WithName(repoName)
	if err != nil {
		return nil, errors.Wrapf(err, "issue getting repository reference for repository name [%s]", repoName)
	}

	repo, err := client.NewRepository(ctx, repoNameRef, endpoint.URL.String(), tr)
	if err != nil {
		return repo, errors.Wrap(err, "issue creating new repository")
	}
	return repo, nil
}

type existingTokenHandler struct {
	token string
}

func (th *existingTokenHandler) Scheme() string {
	return "bearer"
}

func (th *existingTokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", th.token))
	return nil
}
