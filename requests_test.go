package sdp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestRequestsService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/requests/606" {
			t.Errorf("expected /api/v3/requests/606, got %s", r.URL.Path)
		}
		if r.Header.Get("authtoken") != "test-token" {
			t.Errorf("missing or invalid authtoken: %s", r.Header.Get("authtoken"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"response_status": {
				"status_code": 2000,
				"status": "success"
			},
			"request": {
				"id": "606",
				"subject": "Test maintance for multiselect field request",
				"status": {
					"id": "2",
					"name": "Open",
					"color": "#0066ff"
				},
				"udf_fields": {
					"udf_multiselect_8191": [
						{"id": "2106", "name": "user1@agneko.com"}
					]
				}
			}
		}`))
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// 1. Test Default Request (map[string]any)
	req, err := client.Requests.Get(context.Background(), "606")
	if err != nil {
		t.Fatalf("client.Requests.Get failed: %v", err)
	}
	if req.ID != "606" {
		t.Errorf("expected ID 606, got %s", req.ID)
	}
	if req.Subject != "Test maintance for multiselect field request" {
		t.Errorf("unexpected subject: %s", req.Subject)
	}

	// 2. Test Generic Typed GetRequest
	type CustomUDF struct {
		Approvers []sdp.Ref `json:"udf_multiselect_8191"`
	}

	typedReq, err := sdp.GetRequest[CustomUDF](context.Background(), client, "606")
	if err != nil {
		t.Fatalf("sdp.GetRequest failed: %v", err)
	}
	if len(typedReq.UDFFields.Approvers) != 1 || typedReq.UDFFields.Approvers[0].Name != "user1@agneko.com" {
		t.Errorf("unexpected typed UDF: %+v", typedReq.UDFFields)
	}
}

func TestRequestsService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/requests" {
			t.Errorf("expected /api/v3/requests, got %s", r.URL.Path)
		}

		// Verify form-data input_data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		inputData := r.FormValue("input_data")
		if inputData == "" {
			t.Fatalf("input_data parameter was missing from request body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"response_status": {
				"status_code": 2000,
				"status": "success"
			},
			"request": {
				"id": "701",
				"subject": "New Ticket Created"
			}
		}`))
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("token123"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	newTicket := map[string]any{
		"subject": "New Ticket Created",
	}

	created, err := client.Requests.Create(context.Background(), newTicket)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID != "701" {
		t.Errorf("expected ID 701, got %s", created.ID)
	}
}

func TestRequestsService_ErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"response_status": {
				"status_code": 4000,
				"status": "failed",
				"messages": [
					{
						"status_code": 4001,
						"field": "subject",
						"message": "Subject is mandatory"
					}
				]
			}
		}`))
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("token"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.Requests.Get(context.Background(), "999")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	apiErr, ok := err.(*sdp.APIError)
	if !ok {
		t.Fatalf("expected *sdp.APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode != http.StatusBadRequest {
		t.Errorf("expected HTTP 400, got %d", apiErr.HTTPStatusCode)
	}
	if len(apiErr.ResponseStatus) == 0 || apiErr.ResponseStatus[0].StatusCode != 4000 {
		t.Errorf("unexpected ResponseStatus: %+v", apiErr.ResponseStatus)
	}
}
