package sdp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestRequestsService_Worklogs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/worklogs":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": [{"status_code": 2000, "status": "success"}],
				"list_info": {"has_more_rows": false, "start_index": 1, "page": 1, "row_count": 1},
				"worklogs": [{
					"id": "3501",
					"description": "Troubleshot network issue",
					"time_spent": {"hours": 1, "minutes": 30, "value": 5400000},
					"start_time": {"display_value": "Apr 23, 2025 03:19 PM", "value": 1745423155256},
					"owner": {"id": "101", "name": "John Doe"},
					"charge_out_rate": 50.0
				}]
			}`))

		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/worklogs/3501":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"worklog": {
					"id": "3501",
					"description": "Troubleshot network issue",
					"time_spent": {"hours": "1", "minutes": "30", "value": "5400000"}
				}
			}`))

		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/requests/606/worklogs":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			var body map[string]any
			if err := json.Unmarshal([]byte(r.FormValue("input_data")), &body); err != nil {
				t.Fatalf("invalid input_data JSON: %v", err)
			}
			if _, ok := body["worklog"]; !ok {
				t.Errorf("expected 'worklog' wrapper in input_data")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"worklog": {"id": "3502", "description": "Added entry"}
			}`))

		case r.Method == http.MethodPut && r.URL.Path == "/api/v3/requests/606/worklogs/3501":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"worklog": {"id": "3501", "description": "Updated entry"}
			}`))

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v3/requests/606/worklogs/3501":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response_status": {"status_code": 2000, "status": "success"}}`))

		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	// ListWorklogs (verifies numeric TimeSpent & numeric SDPTime)
	listResp, err := client.Requests.ListWorklogs(ctx, "606", nil)
	if err != nil {
		t.Fatalf("ListWorklogs failed: %v", err)
	}
	if len(listResp.Worklogs) != 1 || listResp.Worklogs[0].TimeSpent == nil {
		t.Fatalf("unexpected worklogs: %+v", listResp.Worklogs)
	}
	tsSpent := listResp.Worklogs[0].TimeSpent
	if tsSpent.Hours != "1" || tsSpent.Minutes != "30" || tsSpent.Value != "5400000" {
		t.Errorf("unexpected TimeSpent: %+v", tsSpent)
	}

	// GetWorklog (verifies string TimeSpent)
	wl, err := client.Requests.GetWorklog(ctx, "606", "3501")
	if err != nil {
		t.Fatalf("GetWorklog failed: %v", err)
	}
	if wl.TimeSpent == nil || wl.TimeSpent.Hours != "1" {
		t.Errorf("unexpected GetWorklog TimeSpent: %+v", wl.TimeSpent)
	}

	// CreateWorklog
	created, err := client.Requests.CreateWorklog(ctx, "606", map[string]any{
		"description": "Added entry",
		"time_spent":  map[string]any{"hours": "0", "minutes": "45"},
	})
	if err != nil {
		t.Fatalf("CreateWorklog failed: %v", err)
	}
	if created.ID != "3502" {
		t.Errorf("expected ID 3502, got %s", created.ID)
	}

	// UpdateWorklog
	updated, err := client.Requests.UpdateWorklog(ctx, "606", "3501", map[string]any{"description": "Updated entry"})
	if err != nil {
		t.Fatalf("UpdateWorklog failed: %v", err)
	}
	if updated.Description == nil || *updated.Description != "Updated entry" {
		t.Errorf("unexpected updated description: %v", updated.Description)
	}

	// DeleteWorklog
	if err := client.Requests.DeleteWorklog(ctx, "606", "3501"); err != nil {
		t.Fatalf("DeleteWorklog failed: %v", err)
	}
}
