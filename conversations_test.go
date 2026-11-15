package sdp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestRequestsService_Conversations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/conversations":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": [{"status_code": 2000, "status": "success"}],
				"list_info": {"has_more_rows": false, "start_index": 1, "page": 1, "row_count": 1},
				"conversations": [{
					"id": "11",
					"type": "request_reply",
					"has_attachments": true,
					"show_to_requester": true,
					"sent_time": {"display_value": "29/10/2021 08:12 PM", "value": "1635527559135"},
					"content_url": "/api/v3/requests/606/notifications/11?input_data=%7B%22type%22%3A%22request_reply%22%7D"
				}]
			}`))

		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/conversations/11":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"conversation": {
					"id": "11",
					"type": "request_reply",
					"subject": "Re: Ticket #606",
					"description": "<p>Looking into this</p>"
				}
			}`))

		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/requests/606/notifications/11":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"notification": {
					"id": "11",
					"type": "request_reply",
					"subject": "Re: Ticket #606",
					"description": "<p>Full body</p>",
					"attachments": [{
						"id": "901",
						"name": "screenshot.png",
						"content_type": "image/png",
						"content_url": "/api/v3/requests/606/attachments/901/download"
					}]
				}
			}`))

		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/requests/606/notifications":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			var body map[string]any
			if err := json.Unmarshal([]byte(r.FormValue("input_data")), &body); err != nil {
				t.Fatalf("invalid input_data JSON: %v", err)
			}
			if _, ok := body["notification"]; !ok {
				t.Errorf("expected 'notification' wrapper in input_data")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{
				"response_status": {"status_code": 2000, "status": "success"},
				"notification": {
					"id": "12",
					"subject": "Re: Ticket #606",
					"description": "Reply sent"
				}
			}`))

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

	// ListConversations
	listResp, err := client.Requests.ListConversations(ctx, "606", nil)
	if err != nil {
		t.Fatalf("ListConversations failed: %v", err)
	}
	if len(listResp.Conversations) != 1 || listResp.Conversations[0].ID != "11" {
		t.Fatalf("unexpected conversations: %+v", listResp.Conversations)
	}

	// GetConversationContent using content_url
	content, err := client.Requests.GetConversationContent(ctx, listResp.Conversations[0].ContentURL)
	if err != nil {
		t.Fatalf("GetConversationContent failed: %v", err)
	}
	if content.Notification == nil || len(content.Notification.Attachments) != 1 {
		t.Fatalf("unexpected notification content: %+v", content.Notification)
	}
	if content.Notification.Attachments[0].ContentURL != "/api/v3/requests/606/attachments/901/download" {
		t.Errorf("unexpected attachment content_url: %s", content.Notification.Attachments[0].ContentURL)
	}

	// GetConversation
	conv, err := client.Requests.GetConversation(ctx, "606", "11")
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}
	if conv.Subject != "Re: Ticket #606" {
		t.Errorf("unexpected conversation subject: %s", conv.Subject)
	}

	// SendNotification
	sent, err := client.Requests.SendNotification(ctx, "606", map[string]any{
		"subject":     "Re: Ticket #606",
		"description": "Reply sent",
		"type":        "reply",
	})
	if err != nil {
		t.Fatalf("SendNotification failed: %v", err)
	}
	if sent.ID != "12" {
		t.Errorf("expected ID 12, got %s", sent.ID)
	}
}
