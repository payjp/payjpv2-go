package main

import (
	"os"
	"strings"
	"testing"
)

func TestReplaceIDParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic conversions
		{"basic_customerId", "customerId", "customerID"},
		{"basic_paymentId", "paymentId", "paymentID"},
		{"basic_eventId", "eventId", "eventID"},

		// Compound names
		{"compound_paymentMethodId", "paymentMethodId", "paymentMethodID"},
		{"compound_checkoutSessionId", "checkoutSessionId", "checkoutSessionID"},
		{"compound_paymentMethodConfigurationId", "paymentMethodConfigurationId", "paymentMethodConfigurationID"},

		// Multiple IDs in one string
		{"multiple_ids", "customerId and paymentId", "customerID and paymentID"},
		{"multiple_in_function", "func Foo(customerId, paymentId string)", "func Foo(customerID, paymentID string)"},

		// Should NOT match
		{"no_match_uppercase_start", "Invalid", "Invalid"},
		{"no_match_lowercase_id_only", "id", "id"},
		{"no_match_uppercase_ID", "ID", "ID"},
		{"no_match_identifier", "someIdentifier", "someIdentifier"},
		{"no_match_provide", "provide", "provide"},
		{"no_match_inside_word", "providerId", "providerID"}, // This SHOULD match as it ends with Id

		// Word boundary tests
		{"boundary_parenthesis", "GetCustomer(customerId)", "GetCustomer(customerID)"},
		{"boundary_comma", "customerId, paymentId", "customerID, paymentID"},
		{"boundary_newline", "customerId\npaymentId", "customerID\npaymentID"},
		{"boundary_space", "customerId ", "customerID "},

		// Real-world patterns from generated code
		{"param_declaration", "func GetCustomer(customerId string) error", "func GetCustomer(customerID string) error"},
		{"struct_field", "customerId string `json:\"customer_id\"`", "customerID string `json:\"customer_id\"`"},
		{"variable_usage", "return c.GetCustomer(customerId)", "return c.GetCustomer(customerID)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceIDParams(tt.input)
			if result != tt.expected {
				t.Errorf("replaceIDParams(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReplaceFieldName(t *testing.T) {
	tests := []struct {
		name     string
		oldName  string
		newName  string
		input    string
		expected string
	}{
		// JSON200 -> Data conversions
		{"struct_field_pointer", "JSON200", "Data", "JSON200 *CustomerResponse", "Data *CustomerResponse"},
		{"struct_field_value", "JSON200", "Data", "JSON200 CustomerResponse", "Data CustomerResponse"},
		{"field_access", "JSON200", "Data", "resp.JSON200", "resp.Data"},
		{"field_access_chain", "JSON200", "Data", "resp.JSON200.Name", "resp.Data.Name"},
		{"json_tag", "JSON200", "Data", `json:"JSON200"`, `json:"Data"`},

		// JSON201 -> Data conversions
		{"json201_struct_field", "JSON201", "Data", "JSON201 *CreateResponse", "Data *CreateResponse"},
		{"json201_field_access", "JSON201", "Data", "resp.JSON201", "resp.Data"},

		// Error field conversions
		{"bad_request", "ApplicationproblemJSON400", "BadRequest",
			"ApplicationproblemJSON400 *ErrorResponse", "BadRequest *ErrorResponse"},
		{"unauthorized", "ApplicationproblemJSON401", "Unauthorized",
			"ApplicationproblemJSON401 *ErrorResponse", "Unauthorized *ErrorResponse"},
		{"not_found", "ApplicationproblemJSON404", "NotFound",
			"ApplicationproblemJSON404 *ErrorResponse", "NotFound *ErrorResponse"},
		{"unprocessable", "ApplicationproblemJSON422", "UnprocessableEntity",
			"ApplicationproblemJSON422 *ErrorResponse", "UnprocessableEntity *ErrorResponse"},

		// Should NOT match partial names
		{"no_partial_match", "JSON200", "Data", "JSON2001", "JSON2001"},

		// Multiple occurrences
		{"multiple_occurrences", "JSON200", "Data",
			"if resp.JSON200 != nil { return resp.JSON200.ID }",
			"if resp.Data != nil { return resp.Data.ID }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceFieldName(tt.input, tt.oldName, tt.newName)
			if result != tt.expected {
				t.Errorf("replaceFieldName(%q, %q, %q) = %q, want %q",
					tt.input, tt.oldName, tt.newName, result, tt.expected)
			}
		})
	}
}

func TestExtractErrorFieldMappings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:     "no_error_fields",
			content:  "type Response struct { Data string }",
			expected: map[string]string{},
		},
		{
			name:    "single_error_field",
			content: "ApplicationproblemJSON400 *ErrorResponse",
			expected: map[string]string{
				"ApplicationproblemJSON400": "BadRequest",
			},
		},
		{
			name:    "multiple_error_fields",
			content: "ApplicationproblemJSON400 *ErrorResponse\nApplicationproblemJSON404 *ErrorResponse\nApplicationproblemJSON500 *ErrorResponse",
			expected: map[string]string{
				"ApplicationproblemJSON400": "BadRequest",
				"ApplicationproblemJSON404": "NotFound",
				"ApplicationproblemJSON500": "InternalServerError",
			},
		},
		{
			name:    "duplicates_deduplicated",
			content: "ApplicationproblemJSON400 *ErrorResponse\nApplicationproblemJSON400 *ErrorResponse",
			expected: map[string]string{
				"ApplicationproblemJSON400": "BadRequest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractErrorFieldMappings(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("extractErrorFieldMappings() returned %d mappings, want %d", len(result), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("extractErrorFieldMappings()[%q] = %q, want %q", k, result[k], v)
				}
			}
		})
	}
}

func TestExtractErrorMappings(t *testing.T) {
	errorFieldMappings := map[string]string{
		"ApplicationproblemJSON400": "BadRequest",
		"ApplicationproblemJSON401": "Unauthorized",
		"ApplicationproblemJSON404": "NotFound",
		"ApplicationproblemJSON422": "UnprocessableEntity",
	}

	mappings := extractErrorMappings(errorFieldMappings)

	// Should have 4 error mappings (400, 401, 404, 422)
	if len(mappings) != 4 {
		t.Errorf("extractErrorMappings() returned %d mappings, want 4", len(mappings))
	}

	// Check that mappings are sorted by status code
	for i := 1; i < len(mappings); i++ {
		if mappings[i].StatusCode < mappings[i-1].StatusCode {
			t.Errorf("mappings not sorted: %d comes after %d", mappings[i].StatusCode, mappings[i-1].StatusCode)
		}
	}

	// Check specific mappings
	expected := map[int]string{
		400: "BadRequest",
		401: "Unauthorized",
		404: "NotFound",
		422: "UnprocessableEntity",
	}

	for _, m := range mappings {
		if expected[m.StatusCode] != m.FieldName {
			t.Errorf("mapping for status %d = %q, want %q", m.StatusCode, m.FieldName, expected[m.StatusCode])
		}
	}
}

func TestGenerateErrorMappingsFile(t *testing.T) {
	tmpFile := "test_error_mappings.gen.go"
	defer os.Remove(tmpFile)

	mappings := []ErrorMapping{
		{"BadRequest", 400},
		{"NotFound", 404},
	}

	err := generateErrorMappingsFile(tmpFile, mappings)
	if err != nil {
		t.Fatalf("generateErrorMappingsFile() error = %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	// Check that file contains expected content
	expected := []string{
		"// Code generated by postprocess. DO NOT EDIT.",
		"package payjpv2",
		"import \"net/http\"",
		"type ErrorFieldMapping struct",
		"var ErrorFieldMappings = []ErrorFieldMapping",
		`{"BadRequest", http.StatusBadRequest}`,
		`{"NotFound", http.StatusNotFound}`,
	}

	for _, exp := range expected {
		if !strings.Contains(string(content), exp) {
			t.Errorf("generated file missing expected content: %q", exp)
		}
	}
}

func TestHttpStatusName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{400, "BadRequest"},
		{401, "Unauthorized"},
		{403, "Forbidden"},
		{404, "NotFound"},
		{422, "UnprocessableEntity"},
		{500, "InternalServerError"},
		{999, "999"}, // unknown status code
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := httpStatusName(tt.code)
			if result != tt.expected {
				t.Errorf("httpStatusName(%d) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}
