package acl

import (
	"github.com/stitchfix/flotilla-os/flotilla"
	"net/http"
)

type Client interface {
	AnnotateLaunchRequest(headers http.Header, v2 *flotilla.LaunchRequestV2) error
}
