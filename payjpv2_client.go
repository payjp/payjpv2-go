package payjpv2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

const (
	// BINDINGS_VERSION is the version of the Go SDK bindings, will be set by the Makefile
	BINDINGS_VERSION = "2.0.0"
	// DEFAULT_BASE_URL is the default base URL for the PAY.JP API
	DEFAULT_BASE_URL = "https://api.pay.jp"
)

// clientUserAgent represents the client user agent information
type clientUserAgent struct {
	BindingsVersion string `json:"bindings_version"`
	Lang            string `json:"lang"`
	LangVersion     string `json:"lang_version"`
	Publisher       string `json:"publisher"`
	Uname           string `json:"uname"`
}

// WithUserAgent returns a ClientOption that sets the User-Agent header
func WithUserAgent(userAgent string) ClientOption {
	return WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", userAgent)
		return nil
	})
}

// WithXPayjpClientUserAgent returns a ClientOption that sets the X-Payjp-Client-User-Agent header
func WithXPayjpClientUserAgent(jsonData string) ClientOption {
	return WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-Payjp-Client-User-Agent", jsonData)
		return nil
	})
}

// WithAPIKey returns a ClientOption that sets the Authorization header with the API key
func WithAPIKey(apiKey string) ClientOption {
	return WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		return nil
	})
}

// NewPayjpClientWithResponses creates a new PAY.JP V2 client with request editor function.
func NewPayjpClientWithResponses(apiKey string, opts ...ClientOption) (*ClientWithResponses, error) {
	// Collect system information
	langVersion := runtime.Version()
	uname := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// Create client user agent data
	ua := clientUserAgent{
		BindingsVersion: BINDINGS_VERSION,
		Lang:            "go",
		LangVersion:     langVersion,
		Publisher:       "payjp",
		Uname:           uname,
	}

	// Convert to JSON
	uaJSON, err := json.Marshal(ua)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user agent data: %w", err)
	}

	// Prepend our default options
	defaultOpts := []ClientOption{
		WithUserAgent(fmt.Sprintf("payjp/payjpv2 GoBindings/%s", BINDINGS_VERSION)),
		WithXPayjpClientUserAgent(string(uaJSON)),
		WithAPIKey(apiKey),
	}
	opts = append(defaultOpts, opts...)

	// Create client with default base URL
	client, err := NewClientWithResponses(DEFAULT_BASE_URL, opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}