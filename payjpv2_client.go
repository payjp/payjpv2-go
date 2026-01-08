package payjpv2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

const (
	// BINDINGS_VERSION is the version of the Go SDK bindings, will be set by the Makefile
	BINDINGS_VERSION = "1.0.4"
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

// WithIdempotencyKey returns a RequestEditorFn that sets the Idempotency-Key header
func WithIdempotencyKey(idempotencyKey string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Idempotency-Key", idempotencyKey)
		return nil
	}
}

// NewPayjpClientWithResponses creates a new PAY.JP V2 client with request editor function.
func NewPayjpClientWithResponses(apiKey string, opts ...ClientOption) (*ClientWithResponses, error) {
	// Validate API key
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if !strings.HasPrefix(apiKey, "sk_") {
		return nil, fmt.Errorf("invalid API key format: must start with 'sk_'")
	}

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

// APIError represents an error response from the PAY.JP API.
// It provides structured access to error details returned by the API.
type APIError struct {
	// StatusCode is the HTTP status code of the response
	StatusCode int
	// Body is the parsed error response from the API
	Body *ErrorResponse
	// RawBody is the raw response body bytes
	RawBody []byte
	// Err is the underlying error, if any
	Err error
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.Body != nil {
		if e.Body.Detail != nil && *e.Body.Detail != "" {
			return fmt.Sprintf("PAY.JP API error %d: %s - %s", e.StatusCode, e.Body.Title, *e.Body.Detail)
		}
		return fmt.Sprintf("PAY.JP API error %d: %s", e.StatusCode, e.Body.Title)
	}
	return fmt.Sprintf("PAY.JP API error %d", e.StatusCode)
}

// Unwrap returns the underlying error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsBadRequest returns true if the error is a 400 Bad Request error.
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsUnprocessableEntity returns true if the error is a 422 Unprocessable Entity error.
func (e *APIError) IsUnprocessableEntity() bool {
	return e.StatusCode == http.StatusUnprocessableEntity
}

// ParseAPIError extracts an APIError from a response struct if an error occurred.
// It checks the response for error fields (BadRequest, NotFound, UnprocessableEntity)
// and returns an APIError if one is found, or nil if the request was successful.
//
// Example usage:
//
//	resp, err := client.GetCustomerWithResponse(ctx, customerID)
//	if err != nil {
//	    return err
//	}
//	if apiErr := payjpv2.ParseAPIError(resp); apiErr != nil {
//	    return apiErr
//	}
//	// Use resp.Result for successful response
func ParseAPIError(resp interface{}) *APIError {
	if resp == nil {
		return nil
	}

	v := reflect.ValueOf(resp)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	// Get HTTPResponse to extract status code
	httpRespField := v.FieldByName("HTTPResponse")
	var statusCode int
	if httpRespField.IsValid() && !httpRespField.IsNil() {
		httpResp := httpRespField.Interface().(*http.Response)
		statusCode = httpResp.StatusCode
	}

	// Get raw body
	var rawBody []byte
	bodyField := v.FieldByName("Body")
	if bodyField.IsValid() {
		rawBody = bodyField.Bytes()
	}

	// Check error fields using the generated mappings
	for _, ef := range ErrorFieldMappings {
		field := v.FieldByName(ef.FieldName)
		if field.IsValid() && !field.IsNil() {
			errResp := field.Interface().(*ErrorResponse)
			return &APIError{
				StatusCode: ef.StatusCode,
				Body:       errResp,
				RawBody:    rawBody,
			}
		}
	}

	// Check if status code indicates an error but no specific error field was found
	if statusCode >= 400 {
		return &APIError{
			StatusCode: statusCode,
			RawBody:    rawBody,
		}
	}

	return nil
}

// Extract extracts API errors from a response and returns them as an error.
// This allows handling both network errors and API errors in a single error check.
//
// Example usage:
//
//	resp, err := payjpv2.Extract(client.GetCustomerWithResponse(ctx, customerID))
//	if err != nil {
//	    var apiErr *payjpv2.APIError
//	    if errors.As(err, &apiErr) {
//	        // handle API error
//	        fmt.Println(apiErr.StatusCode, apiErr.Body.Title)
//	    }
//	    return err
//	}
//	customer := resp.Result
func Extract[T any](resp T, err error) (T, error) {
	if err != nil {
		return resp, err
	}
	if apiErr := ParseAPIError(resp); apiErr != nil {
		return resp, apiErr
	}
	return resp, nil
}
