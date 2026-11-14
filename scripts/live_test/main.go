package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/s2gamb/sdp-go"
)

func main() {
	baseURL := os.Getenv("SDP_BASE_URL")
	token := os.Getenv("SDP_TOKEN")
	if baseURL == "" || token == "" {
		log.Fatal("SDP_BASE_URL and SDP_TOKEN environment variables must be set")
	}

	client, err := sdp.NewClient(baseURL, sdp.WithAuthToken(token))
	if err != nil {
		log.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. List requests
	fmt.Println("=== 1. Testing Requests.List ===")
	listResp, err := client.Requests.List(ctx, &sdp.ListOptions{
		Page:     1,
		RowCount: 5,
	})
	if err != nil {
		log.Fatalf("Requests.List failed: %v", err)
	}
	fmt.Printf("List success: %v, count=%d, has_more=%v\n",
		listResp.ResponseStatus.IsSuccess(), len(listResp.Requests), listResp.ListInfo.HasMoreRows)
	for _, r := range listResp.Requests {
		statusName := "<nil>"
		if r.Status != nil {
			statusName = r.Status.Name
		}
		fmt.Printf("  - [%s] %s (Status: %s)\n", r.ID, r.Subject, statusName)
	}

	// 2. Get an existing request (if any)
	var requesterRef map[string]any
	if len(listResp.Requests) > 0 {
		firstID := listResp.Requests[0].ID
		fmt.Printf("\n=== 2. Testing Requests.Get(%s) ===\n", firstID)
		got, err := client.Requests.Get(ctx, firstID)
		if err != nil {
			log.Fatalf("Requests.Get(%s) failed: %v", firstID, err)
		}
		fmt.Printf("Got request #%s: %q (UDFs: %v)\n", got.ID, got.Subject, got.UDFFields)
		if got.Requester != nil && got.Requester.ID != "" {
			requesterRef = map[string]any{"id": got.Requester.ID}
		}
	}

	// 3. Create a request
	fmt.Println("\n=== 3. Testing Requests.Create ===")
	desc := "Automated integration test created by sdp-go"
	createPayload := map[string]any{
		"subject":     fmt.Sprintf("[sdp-go test] %s", time.Now().Format(time.RFC3339)),
		"description": desc,
	}
	if requesterRef != nil {
		createPayload["requester"] = requesterRef
	}
	created, err := client.Requests.Create(ctx, createPayload)
	if err != nil {
		var apiErr *sdp.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("Requests.Create failed: %v\nRawBody: %s", err, apiErr.RawBody)
		}
		log.Fatalf("Requests.Create failed: %v", err)
	}
	fmt.Printf("Created request #%s: %q\n", created.ID, created.Subject)

	// Ensure cleanup even if later steps fail
	defer func() {
		fmt.Printf("\n=== Cleanup: Deleting request #%s ===\n", created.ID)
		if err := client.Requests.Delete(context.Background(), created.ID); err != nil {
			fmt.Printf("Cleanup Delete(%s) failed: %v\n", created.ID, err)
		} else {
			fmt.Printf("Deleted request #%s\n", created.ID)
		}
	}()

	// 4. Get the created request
	fmt.Printf("\n=== 4. Testing Requests.Get(%s) ===\n", created.ID)
	fetched, err := client.Requests.Get(ctx, created.ID)
	if err != nil {
		log.Fatalf("Requests.Get(%s) failed: %v", created.ID, err)
	}
	fmt.Printf("Fetched #%s: subject=%q\n", fetched.ID, fetched.Subject)

	// 5. Update the created request
	fmt.Printf("\n=== 5. Testing Requests.Update(%s) ===\n", created.ID)
	updatedSubject := fetched.Subject + " [updated]"
	updated, err := client.Requests.Update(ctx, created.ID, map[string]any{
		"subject": updatedSubject,
	})
	if err != nil {
		log.Fatalf("Requests.Update(%s) failed: %v", created.ID, err)
	}
	if updated.Subject != updatedSubject {
		log.Fatalf("Expected subject %q, got %q", updatedSubject, updated.Subject)
	}
	fmt.Printf("Updated #%s: subject=%q\n", updated.ID, updated.Subject)

	// 6. Test generic GetRequest with custom struct
	fmt.Printf("\n=== 6. Testing generic sdp.GetRequest[EmptyUDF](%s) ===\n", created.ID)
	type EmptyUDF struct{}
	genericReq, err := sdp.GetRequest[EmptyUDF](ctx, client, created.ID)
	if err != nil {
		log.Fatalf("sdp.GetRequest failed: %v", err)
	}
	fmt.Printf("Generic GetRequest #%s: subject=%q\n", genericReq.ID, genericReq.Subject)

	// 7. Test error handling with non-existent ID
	fmt.Println("\n=== 7. Testing APIError handling (non-existent ID) ===")
	_, err = client.Requests.Get(ctx, "999999999")
	if err == nil {
		log.Fatal("Expected error for non-existent request ID, got nil")
	}
	var apiErr *sdp.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("Received expected APIError (HTTP %d): %v\n", apiErr.HTTPStatusCode, apiErr)
	} else {
		fmt.Printf("Received non-APIError: %v\n", err)
	}

	fmt.Println("\nAll live tests completed!")
}
