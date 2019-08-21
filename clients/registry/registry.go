package registry

import (
	_ "github.com/docker/distribution/manifest"
	"github.com/docker/distribution/reference"
	_ "github.com/docker/libtrust"
	"github.com/nokia/docker-registry-client/registry"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

type Client interface {
	IsImageValid(imageRef string) (bool, error)
}

type registryClient struct {
	tagFetcher ImageTagFetcher
}

type ImageTagFetcher interface {
	TagsForImage(imageRef reference.Reference) ([]string, error)
}

type imageTagFetcher struct {
	registry *registry.Registry
}

func (i imageTagFetcher) TagsForImage(imageRef reference.Reference) ([]string, error) {
	return i.registry.Tags(imageRef.String())
}

func NewRegistryClient(c config.Config) (Client, error) {
	var (
		err error
		rc  registryClient
		tf  imageTagFetcher
	)

	var (
		host     = "https://registry-1.docker.io/"
		username string
		password string
	)

	if c.IsSet("docker.registry_host") {
		host = c.GetString("docker.registry_host")
	}

	if c.IsSet("docker.registry_username") {
		username = c.GetString("docker.registry_username")
	}

	if c.IsSet("docker.registry_host") {
		host = c.GetString("docker.registry_password")
	}

	tf.registry, err = registry.New(host, username, password)

	if err != nil {
		return &rc, errors.Wrap(err, "problem creating registry")
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
