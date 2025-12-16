package payjpv2

import (
	"context"
	"net/http"
	"strings"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
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
			"sk_test_key",
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

func TestWithIdempotencyKey(t *testing.T) {
	t.Run("sets Idempotency-Key header for POST request", func(t *testing.T) {
		mockTransport := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockTransport}

		client, err := NewPayjpClientWithResponses(
			"sk_test_key",
			WithHTTPClient(httpClient),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Make a POST request with idempotency key
		ctx := context.Background()
		idempotencyKey := "test-idempotency-key-12345"
		email := openapi_types.Email("test@example.com")
		body := CreateCustomerJSONRequestBody{
			Email: &email,
		}
		_, _ = client.CreateCustomerWithResponse(ctx, body, WithIdempotencyKey(idempotencyKey))

		// Verify Idempotency-Key header is set
		capturedKey := mockTransport.capturedHeaders.Get("Idempotency-Key")
		if capturedKey != idempotencyKey {
			t.Errorf("Idempotency-Key header incorrect. Got: %s, Expected: %s", capturedKey, idempotencyKey)
		}
	})

	t.Run("different POST requests can have different idempotency keys", func(t *testing.T) {
		mockTransport := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockTransport}

		client, err := NewPayjpClientWithResponses(
			"sk_test_key",
			WithHTTPClient(httpClient),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		ctx := context.Background()
		email := openapi_types.Email("test@example.com")
		body := CreateCustomerJSONRequestBody{
			Email: &email,
		}

		// First request with first idempotency key
		firstKey := "first-idempotency-key"
		_, _ = client.CreateCustomerWithResponse(ctx, body, WithIdempotencyKey(firstKey))

		if mockTransport.capturedHeaders.Get("Idempotency-Key") != firstKey {
			t.Errorf("First request: Idempotency-Key header incorrect. Got: %s, Expected: %s",
				mockTransport.capturedHeaders.Get("Idempotency-Key"), firstKey)
		}

		// Second request with second idempotency key
		secondKey := "second-idempotency-key"
		_, _ = client.CreateCustomerWithResponse(ctx, body, WithIdempotencyKey(secondKey))

		if mockTransport.capturedHeaders.Get("Idempotency-Key") != secondKey {
			t.Errorf("Second request: Idempotency-Key header incorrect. Got: %s, Expected: %s",
				mockTransport.capturedHeaders.Get("Idempotency-Key"), secondKey)
		}
	})
}

func TestNewPayjpClientWithResponses_Validation(t *testing.T) {
	t.Run("rejects empty API key", func(t *testing.T) {
		_, err := NewPayjpClientWithResponses("")
		if err == nil {
			t.Error("Expected error for empty API key, got nil")
		}
		if err.Error() != "API key cannot be empty" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("rejects invalid API key format", func(t *testing.T) {
		_, err := NewPayjpClientWithResponses("invalid_key")
		if err == nil {
			t.Error("Expected error for invalid API key format, got nil")
		}
		if !strings.Contains(err.Error(), "invalid API key format") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("accepts sk_ prefixed API key", func(t *testing.T) {
		mockTransport := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockTransport}

		_, err := NewPayjpClientWithResponses("sk_test_key", WithHTTPClient(httpClient))
		if err != nil {
			t.Errorf("Expected no error for sk_ prefixed key, got: %v", err)
		}
	})
}

func TestAPIError(t *testing.T) {
	t.Run("Error() with body and detail", func(t *testing.T) {
		detail := "Customer not found"
		apiErr := &APIError{
			StatusCode: 404,
			Body: &ErrorResponse{
				Title:  "Not Found",
				Detail: &detail,
				Status: 404,
			},
		}

		expected := "PAY.JP API error 404: Not Found - Customer not found"
		if apiErr.Error() != expected {
			t.Errorf("Expected error message: %s, got: %s", expected, apiErr.Error())
		}
	})

	t.Run("Error() with body but no detail", func(t *testing.T) {
		apiErr := &APIError{
			StatusCode: 400,
			Body: &ErrorResponse{
				Title:  "Bad Request",
				Status: 400,
			},
		}

		expected := "PAY.JP API error 400: Bad Request"
		if apiErr.Error() != expected {
			t.Errorf("Expected error message: %s, got: %s", expected, apiErr.Error())
		}
	})

	t.Run("Error() without body", func(t *testing.T) {
		apiErr := &APIError{
			StatusCode: 500,
		}

		expected := "PAY.JP API error 500"
		if apiErr.Error() != expected {
			t.Errorf("Expected error message: %s, got: %s", expected, apiErr.Error())
		}
	})

	t.Run("IsNotFound", func(t *testing.T) {
		apiErr := &APIError{StatusCode: 404}
		if !apiErr.IsNotFound() {
			t.Error("Expected IsNotFound() to return true for 404 status")
		}

		apiErr2 := &APIError{StatusCode: 400}
		if apiErr2.IsNotFound() {
			t.Error("Expected IsNotFound() to return false for 400 status")
		}
	})

	t.Run("IsBadRequest", func(t *testing.T) {
		apiErr := &APIError{StatusCode: 400}
		if !apiErr.IsBadRequest() {
			t.Error("Expected IsBadRequest() to return true for 400 status")
		}

		apiErr2 := &APIError{StatusCode: 404}
		if apiErr2.IsBadRequest() {
			t.Error("Expected IsBadRequest() to return false for 404 status")
		}
	})

	t.Run("IsUnprocessableEntity", func(t *testing.T) {
		apiErr := &APIError{StatusCode: 422}
		if !apiErr.IsUnprocessableEntity() {
			t.Error("Expected IsUnprocessableEntity() to return true for 422 status")
		}

		apiErr2 := &APIError{StatusCode: 400}
		if apiErr2.IsUnprocessableEntity() {
			t.Error("Expected IsUnprocessableEntity() to return false for 400 status")
		}
	})
}

func TestParseAPIError(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		result := ParseAPIError(nil)
		if result != nil {
			t.Error("Expected nil for nil input")
		}
	})

	t.Run("returns nil for non-struct input", func(t *testing.T) {
		result := ParseAPIError("string")
		if result != nil {
			t.Error("Expected nil for non-struct input")
		}
	})

	t.Run("returns APIError for response with NotFound", func(t *testing.T) {
		detail := "Customer not found"
		resp := &GetCustomerResponse{
			HTTPResponse: &http.Response{StatusCode: 404},
			NotFound: &ErrorResponse{
				Title:  "Not Found",
				Detail: &detail,
				Status: 404,
			},
		}

		apiErr := ParseAPIError(resp)
		if apiErr == nil {
			t.Fatal("Expected APIError, got nil")
		}
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected status code 404, got: %d", apiErr.StatusCode)
		}
		if !apiErr.IsNotFound() {
			t.Error("Expected IsNotFound() to return true")
		}
	})

	t.Run("returns APIError for response with BadRequest", func(t *testing.T) {
		resp := &GetAllCustomersResponse{
			HTTPResponse: &http.Response{StatusCode: 400},
			BadRequest: &ErrorResponse{
				Title:  "Bad Request",
				Status: 400,
			},
		}

		apiErr := ParseAPIError(resp)
		if apiErr == nil {
			t.Fatal("Expected APIError, got nil")
		}
		if apiErr.StatusCode != 400 {
			t.Errorf("Expected status code 400, got: %d", apiErr.StatusCode)
		}
		if !apiErr.IsBadRequest() {
			t.Error("Expected IsBadRequest() to return true")
		}
	})

	t.Run("returns nil for successful response", func(t *testing.T) {
		resp := &GetCustomerResponse{
			HTTPResponse: &http.Response{StatusCode: 200},
			Result:       &CustomerResponse{},
		}

		apiErr := ParseAPIError(resp)
		if apiErr != nil {
			t.Errorf("Expected nil for successful response, got: %v", apiErr)
		}
	})
}