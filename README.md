# sdp-go

`sdp-go` is an idiomatic, extensible Go client library for the **ManageEngine ServiceDesk Plus (SDP) REST API v3**.

## Features

- **Extensible via Generics (`Request[UDF]`)**: Fully type-safe custom user-defined fields (`udf_fields`) without forking or maintaining separate clients per instance.
- **Dynamic Fallback**: Works out-of-the-box with `map[string]any` when custom field schemas are unknown or dynamic.
- **Service-Oriented Client**: Cleanly organized resource services (`client.Requests.Get`, `client.Requests.List`, etc.).
- **Transparent `input_data` Encoding**: Automatically manages SDP's URL-encoded form data JSON wrapper for POST/PUT requests.
- **Robust Error Handling**: Handles SDP's array vs object `response_status` envelopes and provides structured `APIError` messages.
- **Full `context.Context` Support**: First-class timeout and cancellation support on every API call.

---

## Installation

```bash
go get github.com/s2gamb/sdp-go
```

---

## Quick Start

### 1. Simple Usage (Dynamic Custom Fields)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/s2gamb/sdp-go"
)

func main() {
	client, err := sdp.NewClient("https://sdp.company.com",
		sdp.WithAuthToken("YOUR_TECHNICIAN_API_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get a request by ID
	ticket, err := client.Requests.Get(ctx, "606")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Ticket #%s: %s (Status: %s)\n", ticket.ID, ticket.Subject, ticket.Status.Name)

	// Access custom fields dynamically:
	fmt.Println("Custom fields:", ticket.UDFFields)
}
```

---

### 2. Strongly-Typed Custom Fields (Generics)

If your SDP instance has custom user-defined fields (under `udf_fields`), you can define your own struct:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/s2gamb/sdp-go"
)

// Define your instance's custom fields matching SDP JSON keys:
type CorpUDF struct {
	Approvers []sdp.Ref `json:"udf_multiselect_8191"`
	CostCenter string   `json:"udf_char1"`
}

func main() {
	client, _ := sdp.NewClient("https://sdp.company.com", sdp.WithAuthToken("KEY"))

	// Pass your custom struct in square brackets [CorpUDF]:
	ticket, err := sdp.GetRequest[CorpUDF](context.Background(), client, "606")
	if err != nil {
		log.Fatal(err)
	}

	// Full IDE autocomplete and compile-time type safety!
	for _, approver := range ticket.UDFFields.Approvers {
		fmt.Printf("Approver: %s (ID: %s)\n", approver.Name, approver.ID)
	}
}
```

---

### 3. Listing Requests with Pagination

```go
resp, err := client.Requests.List(ctx, &sdp.ListOptions{
    Page:     1,
    RowCount: 25,
})
if err != nil {
    log.Fatal(err)
}

for _, req := range resp.Requests {
    fmt.Printf("[%s] %s\n", req.ID, req.Subject)
}

if resp.ListInfo.HasMoreRows {
    fmt.Println("More pages available!")
}
```

---

### 4. Authentication Options

```go
// Technician Key (authtoken header)
client, _ := sdp.NewClient("https://sdp.company.com", sdp.WithAuthToken("KEY"))

// OAuth Bearer Token (Authorization: Bearer <token>)
client, _ := sdp.NewClient("https://sdp.company.com", sdp.WithOAuthToken("TOKEN"))

// Zoho OAuth (Authorization: Zoho-oauthtoken <token>)
client, _ := sdp.NewClient("https://sdp.company.com", sdp.WithZohoOAuthToken("TOKEN"))
```

---

## Running Tests

```bash
go test -v ./...
```
