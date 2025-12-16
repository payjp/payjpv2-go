# PAY.JP Go SDK v2

Go SDK for the PAY.JP v2 API, generated using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

## Installation

```bash
go get github.com/payjp/payjpv2-go
```

## Usage

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"

    payjpv2 "github.com/payjp/payjpv2-go"
)

func main() {
    client, err := payjpv2.NewPayjpClientWithResponses(os.Getenv("PAYJP_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get all customers
    resp, err := payjpv2.Extract(client.GetAllCustomersWithResponse(ctx, &payjpv2.GetAllCustomersParams{
        Limit: payjpv2.Int(10),
    }))
    if err != nil {
        var apiErr *payjpv2.APIError
        if errors.As(err, &apiErr) {
            log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body.Title)
        }
        log.Fatal(err)
    }

    fmt.Printf("Found %d customers\n", len(resp.Result.Data))
}
```

## Idempotency Keys

To ensure safe retries of requests, you can use idempotency keys. The `WithIdempotencyKey` function allows you to set an idempotency key on a per-request basis:

```go
import (
    "context"
    "github.com/google/uuid"
    payjpv2 "github.com/payjp/payjpv2-go"
    openapi_types "github.com/oapi-codegen/runtime/types"
)

func createCustomerWithIdempotency() {
    client, _ := payjpv2.NewPayjpClientWithResponses(os.Getenv("PAYJP_API_KEY"))
    ctx := context.Background()

    // Generate a unique idempotency key
    idempotencyKey := uuid.New().String()

    // Create customer with idempotency key
    email := openapi_types.Email("customer@example.com")
    resp, err := client.CreateCustomerWithResponse(
        ctx,
        payjpv2.CreateCustomerJSONRequestBody{
            Email: &email,
        },
        payjpv2.WithIdempotencyKey(idempotencyKey),
    )

    // If the request is retried with the same idempotency key,
    // PAY.JP will return the same response without creating a duplicate
}
```

The idempotency key should be unique for each distinct operation. If a request fails due to network issues, you can safely retry it with the same idempotency key.

## Working with Union Types

This SDK handles discriminated unions for payment methods:

```go
// Create a card payment method
cardRequest := payjpv2.PaymentMethodCreateRequest{}
cardData := payjpv2.PaymentMethodCardCreateRequest{
    Type: "card",
    Card: payjpv2.PaymentMethodCardCreateRequestCard{
        Number:   "4242424242424242",
        ExpMonth: 12,
        ExpYear:  2025,
        Cvc:      "123",
    },
}
err := cardRequest.FromPaymentMethodCardCreateRequest(cardData)

// Create a PayPay payment method
paypayRequest := payjpv2.PaymentMethodCreateRequest{}
paypayData := payjpv2.PaymentMethodPayPayCreateRequest{
    Type: "paypay",
    // PayPay-specific fields.
}
err = paypayRequest.FromPaymentMethodPayPayCreateRequest(paypayData)
```

## Features

- Discriminated union support (oneOf/anyOf with discriminator)
- Type-safe request and response handling
- Support for all PAY.JP v2 API endpoints

## Requirements

- Go 1.21+

## Documentation

- [PAY.JP v2 Documents](https://docs.pay.jp/v2)

## License

MIT
