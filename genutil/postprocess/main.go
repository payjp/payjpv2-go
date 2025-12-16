package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Field name mappings for success response structs
// Error field mappings are generated dynamically from client.gen.go
var fieldMappings = map[string]string{
	"JSON200": "Result",
	"JSON201": "Result",
}

// ErrorMapping represents a mapping from error field name to HTTP status code
type ErrorMapping struct {
	FieldName  string
	StatusCode int
}

// extractErrorFieldMappings dynamically extracts error field mappings from the generated code.
// It finds all ApplicationproblemJSON{XXX} patterns and generates mappings using http.StatusText.
func extractErrorFieldMappings(content string) map[string]string {
	pattern := regexp.MustCompile(`ApplicationproblemJSON(\d+)`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	mappings := make(map[string]string)
	for _, match := range matches {
		oldName := match[0] // e.g., "ApplicationproblemJSON400"
		if _, exists := mappings[oldName]; exists {
			continue
		}
		statusCode, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		newName := httpStatusName(statusCode) // e.g., "BadRequest"
		mappings[oldName] = newName
	}
	return mappings
}

func main() {
	inputFile := "client.gen.go"
	outputMappingsFile := "error_mappings.gen.go"

	// Read the generated file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	content := string(data)

	// Dynamically extract error field mappings from the generated code
	errorFieldMappings := extractErrorFieldMappings(content)

	modified := content

	// Apply success response field name mappings
	for old, new := range fieldMappings {
		modified = replaceFieldName(modified, old, new)
	}

	// Apply dynamically extracted error field mappings
	for old, new := range errorFieldMappings {
		modified = replaceFieldName(modified, old, new)
	}

	// Apply dynamic ID parameter mappings (xxxId -> xxxID)
	modified = replaceIDParams(modified)

	// Write the modified file
	if err := os.WriteFile(inputFile, []byte(modified), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	// Generate error_mappings.gen.go
	errorMappings := extractErrorMappings(errorFieldMappings)
	if err := generateErrorMappingsFile(outputMappingsFile, errorMappings); err != nil {
		fmt.Printf("Error generating %s: %v\n", outputMappingsFile, err)
		os.Exit(1)
	}

	fmt.Println("Successfully post-processed client.gen.go")
	fmt.Printf("Successfully generated %s\n", outputMappingsFile)
	printSummary(content, modified, errorFieldMappings)
}

// replaceFieldName replaces struct field names and their references
func replaceFieldName(content, oldName, newName string) string {
	// Replace struct field declarations (e.g., "JSON200 *CustomerResponse")
	// Pattern: field name followed by type
	fieldPattern := regexp.MustCompile(`\b` + oldName + `(\s+\*?\w+)`)
	content = fieldPattern.ReplaceAllString(content, newName+"$1")

	// Replace field access (e.g., "resp.JSON200")
	accessPattern := regexp.MustCompile(`\.` + oldName + `\b`)
	content = accessPattern.ReplaceAllString(content, "."+newName)

	// Replace in json tags (e.g., `json:"JSON200"`)
	jsonTagPattern := regexp.MustCompile(`json:"` + oldName + `"`)
	content = jsonTagPattern.ReplaceAllString(content, `json:"`+newName+`"`)

	return content
}

// replaceIDParams dynamically replaces ID parameter names to follow Go naming conventions.
// It converts camelCase "xxxId" patterns to "xxxID" (e.g., customerId -> customerID).
// This automatically handles any ID parameters from the OpenAPI spec without manual mapping.
func replaceIDParams(content string) string {
	// Pattern: lowercase letter followed by camelCase ending with "Id"
	// Examples: customerId, paymentFlowId, checkoutSessionId
	// This won't match: Invalid (starts with uppercase), id (no prefix)
	pattern := regexp.MustCompile(`\b([a-z][a-zA-Z]*)Id\b`)
	return pattern.ReplaceAllString(content, "${1}ID")
}

// printSummary prints a summary of changes made
func printSummary(original, modified string, errorFieldMappings map[string]string) {
	if original == modified {
		fmt.Println("No changes were made.")
		return
	}

	fmt.Println("\nChanges applied:")

	// Print success response mappings
	for old, new := range fieldMappings {
		oldCount := strings.Count(original, old)
		if oldCount > 0 {
			fmt.Printf("  - %s → %s: %d replacements\n", old, new, oldCount)
		}
	}

	// Print error field mappings
	for old, new := range errorFieldMappings {
		oldCount := strings.Count(original, old)
		if oldCount > 0 {
			fmt.Printf("  - %s → %s: %d replacements\n", old, new, oldCount)
		}
	}

	// Count dynamic ID replacements (xxxId -> xxxID)
	idPattern := regexp.MustCompile(`\b([a-z][a-zA-Z]*)Id\b`)
	idMatches := idPattern.FindAllString(original, -1)
	if len(idMatches) > 0 {
		// Count unique ID names
		uniqueIDs := make(map[string]int)
		for _, match := range idMatches {
			uniqueIDs[match]++
		}
		fmt.Printf("  - ID naming convention (xxxId → xxxID): %d replacements\n", len(idMatches))
		for id, count := range uniqueIDs {
			newID := idPattern.ReplaceAllString(id, "${1}ID")
			fmt.Printf("      %s → %s: %d\n", id, newID, count)
		}
	}
}

// extractErrorMappings extracts error mappings from the dynamically generated errorFieldMappings
// Returns a sorted slice of ErrorMapping for consistent output
func extractErrorMappings(errorFieldMappings map[string]string) []ErrorMapping {
	var mappings []ErrorMapping
	for old, new := range errorFieldMappings {
		// Extract status code from field name (e.g., "ApplicationproblemJSON400" -> 400)
		statusStr := strings.TrimPrefix(old, "ApplicationproblemJSON")
		statusCode, err := strconv.Atoi(statusStr)
		if err != nil {
			continue
		}
		mappings = append(mappings, ErrorMapping{
			FieldName:  new,
			StatusCode: statusCode,
		})
	}
	// Sort by status code for consistent output
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].StatusCode < mappings[j].StatusCode
	})
	return mappings
}

// generateErrorMappingsFile generates the error_mappings.gen.go file
func generateErrorMappingsFile(filename string, mappings []ErrorMapping) error {
	var sb strings.Builder
	sb.WriteString("// Code generated by postprocess. DO NOT EDIT.\n\n")
	sb.WriteString("package payjpv2\n\n")
	sb.WriteString("import \"net/http\"\n\n")
	sb.WriteString("// ErrorFieldMapping defines the mapping between error field name and HTTP status code\n")
	sb.WriteString("type ErrorFieldMapping struct {\n")
	sb.WriteString("\tFieldName  string\n")
	sb.WriteString("\tStatusCode int\n")
	sb.WriteString("}\n\n")
	sb.WriteString("// ErrorFieldMappings is the list of error field mappings used by ParseAPIError\n")
	sb.WriteString("var ErrorFieldMappings = []ErrorFieldMapping{\n")
	for _, m := range mappings {
		sb.WriteString(fmt.Sprintf("\t{%q, http.Status%s},\n", m.FieldName, httpStatusName(m.StatusCode)))
	}
	sb.WriteString("}\n")

	return os.WriteFile(filename, []byte(sb.String()), 0644)
}

// httpStatusName returns the Go http package constant name for a status code
// It uses http.StatusText and removes spaces (e.g., "Bad Request" -> "BadRequest")
func httpStatusName(code int) string {
	text := http.StatusText(code)
	if text == "" {
		return fmt.Sprintf("%d", code)
	}
	// "Bad Request" -> "BadRequest"
	return strings.ReplaceAll(text, " ", "")
}
