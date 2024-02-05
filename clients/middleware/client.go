package middleware

import (
	"github.com/stitchfix/flotilla-os/state"
	"net/http"
)

type Client interface {
	AnnotateLaunchRequest(headers http.Header, lr *state.LaunchRequestV2) error
}
