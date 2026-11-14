package sdp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestTechniciansService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/technicians/42" {
			t.Errorf("expected /api/v3/technicians/42, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"response_status": {
				"status_code": 2000,
				"status": "success"
			},
			"technician": {
				"id": "42",
				"name": "Alice Tech",
				"email_id": "alice@company.com",
				"login_name": "alice",
				"status": "ACTIVE",
				"udf_fields": {
					"udf_char1": "987654321"
				}
			}
		}`))
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	tech, err := client.Technicians.Get(context.Background(), "42")
	if err != nil {
		t.Fatalf("client.Technicians.Get failed: %v", err)
	}
	if tech.ID != "42" || tech.Name != "Alice Tech" {
		t.Errorf("unexpected technician: %+v", tech)
	}
	if tech.UDFFields["udf_char1"] != "987654321" {
		t.Errorf("unexpected udf_fields: %+v", tech.UDFFields)
	}

	type TechUDF struct {
		TelegramID string `json:"udf_char1"`
	}
	typedTech, err := sdp.GetTechnician[TechUDF](context.Background(), client, "42")
	if err != nil {
		t.Fatalf("sdp.GetTechnician failed: %v", err)
	}
	if typedTech.UDFFields.TelegramID != "987654321" {
		t.Errorf("expected TelegramID 987654321, got %q", typedTech.UDFFields.TelegramID)
	}
}

func TestTechniciansService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/technicians" {
			t.Errorf("expected /api/v3/technicians, got %s", r.URL.Path)
		}

		inputData := r.URL.Query().Get("input_data")
		if inputData == "" {
			t.Fatalf("expected input_data query param")
		}
		var decoded struct {
			ListInfo sdp.ListOptions `json:"list_info"`
		}
		if err := json.Unmarshal([]byte(inputData), &decoded); err != nil {
			t.Fatalf("failed to decode input_data: %v", err)
		}
		if decoded.ListInfo.RowCount != 10 || decoded.ListInfo.Page != 1 {
			t.Errorf("unexpected list_info: %+v", decoded.ListInfo)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"response_status": [{"status_code": 2000, "status": "success"}],
			"list_info": {
				"has_more_rows": false,
				"start_index": 1,
				"page": 1,
				"row_count": 1
			},
			"technicians": [
				{
					"id": "42",
					"name": "Alice Tech",
					"udf_fields": {"udf_char1": "987654321"}
				}
			]
		}`))
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.Technicians.List(context.Background(), &sdp.ListOptions{
		Page:     1,
		RowCount: 10,
	})
	if err != nil {
		t.Fatalf("client.Technicians.List failed: %v", err)
	}
	if len(resp.Technicians) != 1 || resp.Technicians[0].ID != "42" {
		t.Errorf("unexpected technicians list: %+v", resp.Technicians)
	}
}

func TestTechniciansService_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/technicians/42" {
			t.Errorf("expected /api/v3/technicians/42, got %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.FormValue("input_data") == "" {
			t.Fatalf("expected input_data in form body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"response_status": {"status_code": 2000, "status": "success"},
			"technician": {
				"id": "42",
				"name": "Alice Tech",
				"udf_fields": {"udf_char1": "111222333"}
			}
		}`))
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	updated, err := client.Technicians.Update(context.Background(), "42", map[string]any{
		"udf_fields": map[string]any{"udf_char1": "111222333"},
	})
	if err != nil {
		t.Fatalf("client.Technicians.Update failed: %v", err)
	}
	if updated.UDFFields["udf_char1"] != "111222333" {
		t.Errorf("unexpected updated udf_fields: %+v", updated.UDFFields)
	}
}

func TestFindTechnicianByField(t *testing.T) {
	type TechUDF struct {
		TelegramID string `json:"udf_char1"`
	}

	t.Run("exact match", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var decoded struct {
				ListInfo sdp.ListOptions `json:"list_info"`
			}
			if err := json.Unmarshal([]byte(r.URL.Query().Get("input_data")), &decoded); err != nil {
				t.Fatalf("failed to decode input_data: %v", err)
			}
			if decoded.ListInfo.RowCount != 2 {
				t.Errorf("expected row_count 2, got %d", decoded.ListInfo.RowCount)
			}
			if len(decoded.ListInfo.SearchCriteria) != 1 {
				t.Fatalf("expected 1 search criterion, got %d", len(decoded.ListInfo.SearchCriteria))
			}
			sc := decoded.ListInfo.SearchCriteria[0]
			if sc.Field != "udf_fields.udf_char1" || sc.Condition != "is" || sc.Value != "987654321" {
				t.Errorf("unexpected search criterion: %+v", sc)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"response_status": [{"status_code": 2000, "status": "success"}],
				"list_info": {"has_more_rows": false, "row_count": 1},
				"technicians": [
					{"id": "42", "name": "Alice Tech", "udf_fields": {"udf_char1": "987654321"}}
				]
			}`))
		}))
		defer ts.Close()

		client, _ := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
		tech, err := sdp.FindTechnicianByField[TechUDF](context.Background(), client, "udf_fields.udf_char1", "987654321")
		if err != nil {
			t.Fatalf("FindTechnicianByField failed: %v", err)
		}
		if tech.ID != "42" || tech.UDFFields.TelegramID != "987654321" {
			t.Errorf("unexpected technician: %+v", tech)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"response_status": [{"status_code": 2000, "status": "success"}],
				"list_info": {"has_more_rows": false, "row_count": 0},
				"technicians": []
			}`))
		}))
		defer ts.Close()

		client, _ := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
		_, err := client.Technicians.FindByField(context.Background(), "udf_fields.udf_char1", "000")
		if !errors.Is(err, sdp.ErrTechnicianNotFound) {
			t.Errorf("expected ErrTechnicianNotFound, got %v", err)
		}
	})

	t.Run("multiple found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"response_status": [{"status_code": 2000, "status": "success"}],
				"list_info": {"has_more_rows": false, "row_count": 2},
				"technicians": [
					{"id": "42", "name": "Alice Tech"},
					{"id": "43", "name": "Bob Tech"}
				]
			}`))
		}))
		defer ts.Close()

		client, _ := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
		_, err := sdp.FindTechnicianByField[TechUDF](context.Background(), client, "udf_fields.udf_char1", "987654321")
		if !errors.Is(err, sdp.ErrMultipleTechniciansFound) {
			t.Errorf("expected ErrMultipleTechniciansFound, got %v", err)
		}
	})
}
