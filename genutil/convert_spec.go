package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// Read the OpenAPI spec
	data, err := os.ReadFile("../openapi.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Change OpenAPI version from 3.1.0 to 3.0.3 for better compatibility
	spec["openapi"] = "3.0.3"

	// Remove or fix problematic null types and anyOf
	fixNullTypes(spec)
	
	// Change application/problem+json to application/json
	fixContentTypes(spec)

	// Write the modified spec
	modifiedData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("openapi-converted.json", modifiedData, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully converted OpenAPI spec for oapi-codegen compatibility")
}

func fixNullTypes(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == "anyOf" {
				if arr, ok := value.([]interface{}); ok {
					// Remove null types from anyOf arrays
					var newArr []interface{}
					hasNull := false
					for _, item := range arr {
						if m, ok := item.(map[string]interface{}); ok {
							if typeVal, exists := m["type"]; exists && typeVal == "null" {
								hasNull = true
								continue
							}
						}
						newArr = append(newArr, item)
					}
					if len(newArr) == 1 {
						// If only one type remains, replace anyOf with the single type's properties
						if m, ok := newArr[0].(map[string]interface{}); ok {
							delete(v, "anyOf")
							for k, val := range m {
								v[k] = val
							}
							// If nullable, add it as a property
							if hasNull {
								v["nullable"] = true
							}
						}
					} else if len(newArr) > 1 {
						v[key] = newArr
					} else if len(newArr) == 0 && hasNull {
						// If only null type existed, make it nullable string
						delete(v, "anyOf")
						v["type"] = "string"
						v["nullable"] = true
					}
				}
			} else if key == "type" {
				// Fix array types that include null
				if arr, ok := value.([]interface{}); ok {
					var nonNullTypes []interface{}
					hasNull := false
					for _, t := range arr {
						if t == "null" {
							hasNull = true
						} else {
							nonNullTypes = append(nonNullTypes, t)
						}
					}
					if len(nonNullTypes) == 1 {
						v["type"] = nonNullTypes[0]
						if hasNull {
							v["nullable"] = true
						}
					} else if len(nonNullTypes) > 1 {
						v["type"] = nonNullTypes
					}
				}
			} else {
				fixNullTypes(value)
			}
		}
	case []interface{}:
		for _, item := range v {
			fixNullTypes(item)
		}
	}
}

func fixContentTypes(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == "content" {
				if content, ok := value.(map[string]interface{}); ok {
					// Check if application/problem+json exists
					if problemJson, exists := content["application/problem+json"]; exists {
						// Copy to application/json
						content["application/json"] = problemJson
						// Remove application/problem+json
						delete(content, "application/problem+json")
					}
				}
			}
			fixContentTypes(value)
		}
	case []interface{}:
		for _, item := range v {
			fixContentTypes(item)
		}
	}
}