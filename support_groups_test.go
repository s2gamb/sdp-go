package sdp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestSupportGroupsService_ListAndGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/support_groups":
			_, _ = w.Write([]byte(`{
				"response_status":[{"status_code":2000,"status":"success"}],
				"list_info":{"has_more_rows":false,"start_index":1,"row_count":2},
				"support_groups":[
					{"id":"301","name":"Входящие","description":null,"deleted":false},
					{"id":"601","name":"HELPDESK","description":"Support","deleted":false}
				]
			}`))
		case "/api/v3/support_groups/301":
			_, _ = w.Write([]byte(`{
				"response_status":{"status_code":2000,"status":"success"},
				"support_group":{
					"id":"301",
					"name":"Входящие",
					"deleted":false,
					"technicians":[
						{"id":"42","name":"Alice Tech"},
						{"id":"99","name":"Bob Tech"}
					]
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client, err := sdp.NewClient(srv.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	listResp, err := client.SupportGroups.List(ctx, &sdp.ListOptions{RowCount: 50})
	if err != nil {
		t.Fatalf("SupportGroups.List failed: %v", err)
	}
	if len(listResp.SupportGroups) != 2 || listResp.SupportGroups[0].ID != "301" {
		t.Fatalf("unexpected support_groups: %+v", listResp.SupportGroups)
	}

	grp, err := client.SupportGroups.Get(ctx, "301")
	if err != nil {
		t.Fatalf("SupportGroups.Get failed: %v", err)
	}
	if grp.ID != "301" || len(grp.Technicians) != 2 || grp.Technicians[0].ID != "42" {
		t.Fatalf("unexpected support_group detail: %+v", grp)
	}
}
