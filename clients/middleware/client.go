package middleware

import (
	"github.com/stitchfix/flotilla-os/state"
	"net/http"
)

type Client interface {
	AnnotateLaunchRequest(headers http.Header, lr *state.LaunchRequestV2) error
}

type MiddlewareClient struct{}

func NewClient() (Client, error) {
	return &MiddlewareClient{}, nil
}

func (mwC *MiddlewareClient) AnnotateLaunchRequest(headers http.Header, lr *state.LaunchRequestV2) error {
	return nil
}
