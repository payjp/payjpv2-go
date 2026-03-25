package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	payjpv2 "github.com/payjp/payjpv2-go"
)

// Helper functions for creating metadata values
func createMetadataString(v string) payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties {
	var prop payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties
	_ = prop.FromCustomerCreateRequestMetadata0(v)
	return prop
}

func createMetadataInt(v int) payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties {
	var prop payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties
	_ = prop.FromCustomerCreateRequestMetadata1(v)
	return prop
}

func createMetadataBool(v bool) payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties {
	var prop payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties
	_ = prop.FromCustomerCreateRequestMetadata2(v)
	return prop
}

func updateMetadataString(v string) payjpv2.CustomerUpdateRequest_Metadata_AdditionalProperties {
	var prop payjpv2.CustomerUpdateRequest_Metadata_AdditionalProperties
	_ = prop.FromCustomerUpdateRequestMetadata0(v)
	return prop
}

func updateMetadataInt(v int) payjpv2.CustomerUpdateRequest_Metadata_AdditionalProperties {
	var prop payjpv2.CustomerUpdateRequest_Metadata_AdditionalProperties
	_ = prop.FromCustomerUpdateRequestMetadata1(v)
	return prop
}

func main() {
	// Get settings from environment variables
	apiHost := os.Getenv("PAYJP_API_HOST")
	if apiHost == "" {
		apiHost = "https://api.pay.jp"
	}
	apiKey := os.Getenv("PAYJP_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: Please set the PAYJP_API_KEY environment variable")
	}

	// Initialize the PAY.JP client
	client, err := payjpv2.NewPayjpClientWithResponses(apiKey, payjpv2.WithBaseURL(apiHost))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PAY.JP Go SDK (oapi-codegen) initialized successfully")

	ctx := context.Background()

	// Example 1: Create a customer with idempotency key
	idempotencyKey := uuid.New().String()
	fmt.Printf("Using Idempotency-Key: %s\n", idempotencyKey)

	email := types.Email("jennyrosen@example.com")
	metadata := map[string]payjpv2.CustomerCreateRequest_Metadata_AdditionalProperties{
		"key1": createMetadataString("value1"),
		"key2": createMetadataInt(123),
		"key3": createMetadataBool(true),
	}
	customerRequest := payjpv2.CustomerCreateRequest{
		Email:    &email,
		Metadata: &metadata,
	}

	customerResponse, err := payjpv2.Extract(client.CreateCustomerWithResponse(
		ctx,
		customerRequest,
		payjpv2.WithIdempotencyKey(idempotencyKey),
	))
	if err != nil {
		var apiErr *payjpv2.APIError
		if errors.As(err, &apiErr) {
			if apiErr.Body != nil {
				log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
			}
			log.Fatalf("API error %d: %s", apiErr.StatusCode, string(apiErr.RawBody))
		}
		log.Fatal(err)
	}
	fmt.Printf("Created customer: %s\n", customerResponse.Result.Id)
	metadataJSON, _ := json.Marshal(customerResponse.Result.Metadata)
	fmt.Printf("Metadata: %s\n", string(metadataJSON))

	// Example 2: Get Customer
	fmt.Println("\n=== 2. Get Customer ===")
	getResponse, err := payjpv2.Extract(client.GetCustomerWithResponse(ctx, customerResponse.Result.Id))
	if err != nil {
		var apiErr *payjpv2.APIError
		if errors.As(err, &apiErr) {
			if apiErr.Body != nil {
				log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
			}
			log.Fatalf("API error %d: %s", apiErr.StatusCode, string(apiErr.RawBody))
		}
		log.Fatal(err)
	}
	fmt.Printf("Retrieved customer: %s\n", getResponse.Result.Id)
	fmt.Printf("Email: %s\n", *getResponse.Result.Email)
	retrievedMetadataJSON, _ := json.Marshal(getResponse.Result.Metadata)
	fmt.Printf("Metadata: %s\n", string(retrievedMetadataJSON))

	// Example 3: Update Customer
	fmt.Println("\n=== 3. Update Customer ===")
	updatedEmail := types.Email("updated@example.com")
	updatedDescription := "Updated description from Go SDK"
	updateMetadata := map[string]payjpv2.CustomerUpdateRequest_Metadata_AdditionalProperties{
		"key1": updateMetadataString("updated_value"),
		"key4": updateMetadataInt(456),
	}
	updateRequest := payjpv2.CustomerUpdateRequest{
		Email:       &updatedEmail,
		Description: &updatedDescription,
		Metadata:    &updateMetadata,
	}
	updateResponse, err := payjpv2.Extract(client.UpdateCustomerWithResponse(ctx, customerResponse.Result.Id, updateRequest))
	if err != nil {
		var apiErr *payjpv2.APIError
		if errors.As(err, &apiErr) {
			if apiErr.Body != nil {
				log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
			}
			log.Fatalf("API error %d: %s", apiErr.StatusCode, string(apiErr.RawBody))
		}
		log.Fatal(err)
	}
	fmt.Printf("Updated customer: %s\n", updateResponse.Result.Id)
	fmt.Printf("New email: %s\n", *updateResponse.Result.Email)
	fmt.Printf("New description: %s\n", *updateResponse.Result.Description)
	updatedMetadataJSON, _ := json.Marshal(updateResponse.Result.Metadata)
	fmt.Printf("Metadata: %s\n", string(updatedMetadataJSON))

	// Example 4: Create a payment method (card)
	cardRequest := payjpv2.PaymentMethodCreateRequest{}
	billingEmail := "jennyrosen@example.com"
	cardData := payjpv2.PaymentMethodCardCreateRequest{
		Type: "card",
		BillingDetails: payjpv2.PaymentMethodCardBillingDetailsRequest{
			Email: &billingEmail,
		},
		Card: payjpv2.PaymentMethodCreateCardDetailsRequest{
			Number:   "4242424242424242",
			ExpMonth: 12,
			ExpYear:  2030,
			Cvc:      "123",
		},
	}
	err = cardRequest.FromPaymentMethodCardCreateRequest(cardData)
	if err != nil {
		log.Fatal(err)
	}

	pmResponse, err := payjpv2.Extract(client.CreatePaymentMethodWithResponse(ctx, cardRequest))
	if err != nil {
		var apiErr *payjpv2.APIError
		if errors.As(err, &apiErr) {
			if apiErr.Body != nil {
				log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
			}
			log.Fatalf("API error %d: %s", apiErr.StatusCode, string(apiErr.RawBody))
		}
		log.Fatal(err)
	}
	fmt.Printf("Created payment method: %+v\n", pmResponse.Result)
}
