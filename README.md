# PAY.JP Go SDK v2 (oapi-codegen)

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
    "fmt"
    "log"
    "net/http"

    payjpv2 "github.com/payjp/payjpv2-go"
)

func main() {
    // Initialize the client
    client, err := payjpv2.NewPayjpClientWithResponses("YOUR_API_KEY_HERE")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    
    // Example: Get all customers
    response, err := client.GetAllCustomersWithResponse(ctx, &payjpv2.GetAllCustomersParams{
        Limit: payjpv2.Int(10),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    if response.StatusCode() == http.StatusOK {
        fmt.Printf("Found %d customers\n", len(response.JSON200.Data))
    }
}
```

## Working with Union Types

This SDK properly handles discriminated unions for payment methods:

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
    // PayPay-specific fields...
}
err = paypayRequest.FromPaymentMethodPayPayCreateRequest(paypayData)
```

## Features

- Full support for discriminated unions (oneOf/anyOf with discriminator)
- Type-safe request and response handling
- Automatic request/response validation
- Built-in retry and error handling options
- Support for all PAY.JP v2 API endpoints

## Requirements

- Go 1.21 or higher
- oapi-codegen v2.5.0 or higher

## License

MIT