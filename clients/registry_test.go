package clients

import (
	"github.com/docker/distribution/reference"
	"testing"
)

type testTagFetcher struct {
	tags []string
}

func (tf *testTagFetcher) TagsForImage(imageRef reference.Reference) ([]string, error) {
	return tf.tags, nil
}

func TestRegistryClient_IsImageValid(t *testing.T) {
	validTags := []string{
		"A", "B", "C",
	}
	ttf := testTagFetcher{validTags}
	rc := registryClient{&ttf}

	img1 := "foo/bar:baz"
	img2 := "cupcake/sprinkles:C"

	valid, _ := rc.IsImageValid(img1)
	if valid {
		t.Errorf("Image [%s] should not be valid, its tag is not present in %s", img1, validTags)
	}

	valid, _ = rc.IsImageValid(img2)
	if !valid {
		t.Errorf("Image [%s] should be valid, its tag is present in %s", img2, validTags)
	}
}
