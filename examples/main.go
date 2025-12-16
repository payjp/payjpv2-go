package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	payjpv2 "github.com/payjp/payjpv2-go"
)

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
	customerRequest := payjpv2.CustomerCreateRequest{
		Email: &email,
	}

	customerResponse, err := payjpv2.Extract(client.CreateCustomerWithResponse(
		ctx,
		customerRequest,
		payjpv2.WithIdempotencyKey(idempotencyKey),
	))
	if err != nil {
		var apiErr *payjpv2.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
		}
		log.Fatal(err)
	}
	fmt.Printf("Created customer: %+v\n", customerResponse.Result)

	// Example 2: Create a payment method (card)
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
			log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
		}
		log.Fatal(err)
	}
	fmt.Printf("Created payment method: %+v\n", pmResponse.Result)
}
