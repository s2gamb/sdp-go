package sdp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestRequestsService_Tasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/tasks":
			inputData := r.URL.Query().Get("input_data")
			if inputData == "" {
				t.Errorf("expected input_data query param for ListTasks")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": [{"status_code": 2000, "status": "success"}],
				"list_info": {"has_more_rows": false, "start_index": 1, "page": 1, "row_count": 1},
				"tasks": [{
					"id": "5564",
					"title": "Review server logs",
					"status": {"id": "1", "name": "Open"},
					"percent_complete": 25
				}]
			}`))

		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/tasks/5564":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"task": {
					"id": "5564",
					"title": "Review server logs",
					"status": {"id": "1", "name": "Open"},
					"scheduled_start_time": {"display_value": "Feb 25, 2025", "value": 1740470400000},
					"udf_fields": {"udf_char1": "infra"}
				}
			}`))

		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/requests/606/tasks":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			var body map[string]any
			if err := json.Unmarshal([]byte(r.FormValue("input_data")), &body); err != nil {
				t.Fatalf("invalid input_data JSON: %v", err)
			}
			if _, ok := body["task"]; !ok {
				t.Errorf("expected 'task' wrapper in input_data")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"task": {"id": "5565", "title": "New task"}
			}`))

		case r.Method == http.MethodPut && r.URL.Path == "/api/v3/requests/606/tasks/5564":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"task": {"id": "5564", "title": "Updated task"}
			}`))

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v3/requests/606/tasks/5564":
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

	// ListTasks
	listResp, err := client.Requests.ListTasks(ctx, "606", &sdp.ListOptions{Page: 1, RowCount: 10})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(listResp.Tasks) != 1 || listResp.Tasks[0].ID != "5564" {
		t.Errorf("unexpected tasks: %+v", listResp.Tasks)
	}

	// GetTask (default + generic typed UDF)
	task, err := client.Requests.GetTask(ctx, "606", "5564")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if task.ScheduledStartTime == nil || task.ScheduledStartTime.Value != "1740470400000" {
		t.Errorf("unexpected numeric SDPTime: %+v", task.ScheduledStartTime)
	}

	type TaskUDF struct {
		Team string `json:"udf_char1"`
	}
	typedTask, err := sdp.GetRequestTask[TaskUDF](ctx, client, "606", "5564")
	if err != nil {
		t.Fatalf("GetRequestTask failed: %v", err)
	}
	if typedTask.UDFFields.Team != "infra" {
		t.Errorf("unexpected TaskUDF: %+v", typedTask.UDFFields)
	}

	// CreateTask
	created, err := client.Requests.CreateTask(ctx, "606", map[string]any{"title": "New task"})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if created.ID != "5565" {
		t.Errorf("expected ID 5565, got %s", created.ID)
	}

	// UpdateTask
	updated, err := client.Requests.UpdateTask(ctx, "606", "5564", map[string]any{"title": "Updated task"})
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if updated.Title != "Updated task" {
		t.Errorf("expected Updated task, got %s", updated.Title)
	}

	// DeleteTask
	if err := client.Requests.DeleteTask(ctx, "606", "5564"); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}
}
