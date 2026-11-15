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

	// 1. List technicians
	fmt.Println("=== 1. Testing Technicians.List ===")
	listResp, err := client.Technicians.List(ctx, &sdp.ListOptions{
		Page:     1,
		RowCount: 5,
	})
	if err != nil {
		log.Fatalf("Technicians.List failed: %v", err)
	}
	fmt.Printf("List success: %v, count=%d, has_more=%v\n",
		listResp.ResponseStatus.IsSuccess(), len(listResp.Technicians), listResp.ListInfo.HasMoreRows)
	for _, t := range listResp.Technicians {
		email := "<nil>"
		if t.EmailID != nil {
			email = *t.EmailID
		}
		fmt.Printf("  - [%s] %s (login=%q, email=%q, status=%q)\n", t.ID, t.Name, t.LoginName, email, t.Status)
	}

	if len(listResp.Technicians) == 0 {
		log.Fatal("No technicians returned; cannot continue Get/FindByField/Update checks")
	}

	targetID := listResp.Technicians[0].ID

	// 2. Get technician
	fmt.Printf("\n=== 2. Testing Technicians.Get(%s) ===\n", targetID)
	tech, err := client.Technicians.Get(ctx, targetID)
	if err != nil {
		log.Fatalf("Technicians.Get(%s) failed: %v", targetID, err)
	}
	fmt.Printf("Got technician #%s: name=%q, jobtitle=%q, TechUDFs=%v, UserUDFs=%v\n",
		tech.ID, tech.Name, tech.JobTitle, tech.UDFFields, tech.UserUDFFields)

	// 3. Generic GetTechnician
	fmt.Printf("\n=== 3. Testing generic sdp.GetTechnician[EmptyUDF](%s) ===\n", targetID)
	type EmptyUDF struct{}
	genericTech, err := sdp.GetTechnician[EmptyUDF](ctx, client, targetID)
	if err != nil {
		log.Fatalf("sdp.GetTechnician failed: %v", err)
	}
	fmt.Printf("Generic GetTechnician #%s: name=%q\n", genericTech.ID, genericTech.Name)

	// 4. FindByField (exact match)
	fmt.Printf("\n=== 4. Testing Technicians.FindByField(\"name\", %q) ===\n", tech.Name)
	found, err := client.Technicians.FindByField(ctx, "name", tech.Name)
	if err != nil {
		fmt.Printf("FindByField(\"name\", %q) returned error: %v\n", tech.Name, err)
	} else {
		fmt.Printf("Found technician #%s: name=%q\n", found.ID, found.Name)
	}

	if tech.EmailID != nil && *tech.EmailID != "" {
		fmt.Printf("\n=== 4b. Testing Technicians.FindByField(\"email_id\", %q) ===\n", *tech.EmailID)
		foundByEmail, err := client.Technicians.FindByField(ctx, "email_id", *tech.EmailID)
		if err != nil {
			fmt.Printf("FindByField(\"email_id\", %q) returned error: %v\n", *tech.EmailID, err)
		} else {
			fmt.Printf("Found technician by email #%s: name=%q\n", foundByEmail.ID, foundByEmail.Name)
		}
	}

	// 5. FindByField (not found)
	fmt.Println("\n=== 5. Testing Technicians.FindByField (not found) ===")
	_, err = client.Technicians.FindByField(ctx, "name", "__nonexistent_technician_sdp_go__")
	if !errors.Is(err, sdp.ErrTechnicianNotFound) {
		log.Fatalf("Expected ErrTechnicianNotFound, got: %v", err)
	}
	fmt.Printf("Received expected error: %v\n", err)

	// 6. Update technician (description + jobtitle) and restore original values
	fmt.Printf("\n=== 6. Testing Technicians.Update(%s) ===\n", targetID)
	var origDesc string
	if tech.Description != nil {
		origDesc = *tech.Description
	}
	origJobTitle := tech.JobTitle
	testDesc := "sdp-go test description"
	if origDesc == testDesc {
		testDesc = "sdp-go test description 2"
	}
	testJobTitle := "sdp-go test title"
	if origJobTitle == testJobTitle {
		testJobTitle = "sdp-go test title 2"
	}

	updated, err := client.Technicians.Update(ctx, targetID, map[string]any{
		"description": testDesc,
		"jobtitle":    testJobTitle,
	})
	if err != nil {
		log.Fatalf("Technicians.Update(%s) failed: %v", targetID, err)
	}
	var updatedDesc string
	if updated.Description != nil {
		updatedDesc = *updated.Description
	}
	fmt.Printf("Updated #%s description=%q, jobtitle=%q\n", updated.ID, updatedDesc, updated.JobTitle)

	// Restore original description & jobtitle
	restored, err := client.Technicians.Update(ctx, targetID, map[string]any{
		"description": origDesc,
		"jobtitle":    origJobTitle,
	})
	if err != nil {
		log.Fatalf("Failed to restore fields for #%s: %v", targetID, err)
	}
	var restoredDesc string
	if restored.Description != nil {
		restoredDesc = *restored.Description
	}
	fmt.Printf("Restored #%s description=%q, jobtitle=%q\n", restored.ID, restoredDesc, restored.JobTitle)

	// 7. Error handling for non-existent ID
	fmt.Println("\n=== 7. Testing APIError handling (non-existent ID) ===")
	_, err = client.Technicians.Get(ctx, "999999999")
	if err == nil {
		log.Fatal("Expected error for non-existent technician ID, got nil")
	}
	var apiErr *sdp.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("Received expected APIError (HTTP %d): %v\n", apiErr.HTTPStatusCode, apiErr)
	} else {
		fmt.Printf("Received non-APIError: %v\n", err)
	}

	fmt.Println("\nAll live technician tests completed!")
}
