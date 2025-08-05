package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	payjpv2 "github.com/payjp/payjpv2-go"
)

func main() {
	// Initialize the PAY.JP client
	client, err := payjpv2.NewClientWithResponses(
		"https://api.pay.jp",
		payjpv2.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			// Set the Authorization header with your API key
			req.Header.Set("Authorization", "Bearer YOUR_API_KEY_HERE")
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PAY.JP Go SDK (oapi-codegen) initialized successfully")

	// Example: Create a payment method (card)
	ctx := context.Background()
	
	// Create a card payment method
	cardRequest := payjpv2.PaymentMethodCreateRequest{}
	cardData := payjpv2.PaymentMethodCardCreateRequest{
		Type: "card",
		Card: payjpv2.PaymentMethodCreateCardDetailsRequest{
			Number:   "4242424242424242",
			ExpMonth: 12,
			ExpYear:  2025,
			Cvc:      "123",
		},
	}
	err = cardRequest.FromPaymentMethodCardCreateRequest(cardData)
	if err != nil {
		log.Fatal(err)
	}

	// Create the payment method
	response, err := client.CreatePaymentMethodWithResponse(ctx, cardRequest)
	if err != nil {
	    log.Fatal(err)
	}
	if response.StatusCode() != http.StatusOK {
	    log.Fatalf("Failed to create payment method: %s", response.Status())
	}
	fmt.Printf("Created payment method: %+v\n", response.JSON200)

	_ = client
	_ = cardRequest
}