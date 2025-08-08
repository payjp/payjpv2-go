package payjpv2

import (
	"context"
	"net/http"
	"strings"
	"testing"

)

// mockRoundTripper captures headers for testing
type mockRoundTripper struct {
	capturedHeaders http.Header
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedHeaders = req.Header.Clone()
	
	// Return a mock response to avoid actual API call
	return &http.Response{
		StatusCode: 401,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}, nil
}

func TestNewPayjpClientWithResponses(t *testing.T) {
	// Create mock transport to capture headers
	mockTransport := &mockRoundTripper{}
	httpClient := &http.Client{
		Transport: mockTransport,
	}

	// Create client using NewPayjpClientWithResponses
	client, err := NewPayjpClientWithResponses(
		"sk_test_example",
		WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request to trigger header capture
	ctx := context.Background()
	limit := int(1)
	_, _ = client.GetAllCustomersWithResponse(ctx, &GetAllCustomersParams{
		Limit: &limit,
	})

	// Verify User-Agent header
	userAgent := mockTransport.capturedHeaders.Get("User-Agent")
	expectedUserAgentPrefix := "payjp/payjpv2 GoBindings/"
	if !strings.HasPrefix(userAgent, expectedUserAgentPrefix) {
		t.Errorf("User-Agent header incorrect. Got: %s, Expected prefix: %s", userAgent, expectedUserAgentPrefix)
	}

	// Verify X-Payjp-Client-User-Agent header exists and contains expected fields
	clientUserAgent := mockTransport.capturedHeaders.Get("X-Payjp-Client-User-Agent")
	if clientUserAgent == "" {
		t.Error("X-Payjp-Client-User-Agent header is missing")
	}

	// Check that it contains expected JSON fields
	expectedFields := []string{
		`"bindings_version"`,
		`"lang":"go"`,
		`"lang_version"`,
		`"publisher":"payjp"`,
		`"uname"`,
	}
	for _, field := range expectedFields {
		if !strings.Contains(clientUserAgent, field) {
			t.Errorf("X-Payjp-Client-User-Agent missing field: %s. Got: %s", field, clientUserAgent)
		}
	}

	// Verify Authorization header
	auth := mockTransport.capturedHeaders.Get("Authorization")
	if auth != "Bearer sk_test_example" {
		t.Errorf("Authorization header incorrect. Got: %s, Expected: Bearer sk_test_example", auth)
	}
}

func TestClientAPIKeyAuthorization(t *testing.T) {
	// Test that API key is properly set in Authorization header
	mockTransport := &mockRoundTripper{}
	httpClient := &http.Client{Transport: mockTransport}

	apiKey := "sk_test_custom_key_123"
	client, err := NewPayjpClientWithResponses(
		apiKey,
		WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Trigger a request
	ctx := context.Background()
	limit := int(1)
	_, _ = client.GetAllCustomersWithResponse(ctx, &GetAllCustomersParams{
		Limit: &limit,
	})

	// Verify Authorization header contains the API key
	auth := mockTransport.capturedHeaders.Get("Authorization")
	expectedAuth := "Bearer " + apiKey
	if auth != expectedAuth {
		t.Errorf("Authorization header incorrect. Got: %s, Expected: %s", auth, expectedAuth)
	}
}

func TestWithAPIKeyOption(t *testing.T) {
	// Test that WithAPIKey can be used as a standalone option
	mockTransport := &mockRoundTripper{}
	httpClient := &http.Client{Transport: mockTransport}

	apiKey := "sk_test_standalone_key"
	client, err := NewClientWithResponses(
		DEFAULT_BASE_URL,
		WithHTTPClient(httpClient),
		WithAPIKey(apiKey),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create wrapper for response handling
	clientWithResponses := &ClientWithResponses{client}

	// Trigger a request
	ctx := context.Background()
	limit := int(1)
	_, _ = clientWithResponses.GetAllCustomersWithResponse(ctx, &GetAllCustomersParams{
		Limit: &limit,
	})

	// Verify Authorization header
	auth := mockTransport.capturedHeaders.Get("Authorization")
	expectedAuth := "Bearer " + apiKey
	if auth != expectedAuth {
		t.Errorf("Authorization header incorrect. Got: %s, Expected: %s", auth, expectedAuth)
	}
}

func TestClientHeadersIntegration(t *testing.T) {
	// This test shows how headers are properly set in the client
	t.Run("headers are set correctly", func(t *testing.T) {
		mockTransport := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockTransport}

		client, err := NewPayjpClientWithResponses(
			"test_key",
			WithHTTPClient(httpClient),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Trigger a request
		ctx := context.Background()
		limit := int(1)
		_, _ = client.GetAllCustomersWithResponse(ctx, &GetAllCustomersParams{
			Limit: &limit,
		})

		// Verify all required headers are present
		headers := mockTransport.capturedHeaders
		
		if headers.Get("User-Agent") == "" {
			t.Error("User-Agent header is missing")
		}
		if headers.Get("X-Payjp-Client-User-Agent") == "" {
			t.Error("X-Payjp-Client-User-Agent header is missing")
		}
		if headers.Get("Authorization") == "" {
			t.Error("Authorization header is missing")
		}
	})
}