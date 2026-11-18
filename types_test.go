package sdp_test

import (
	"encoding/json"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestFlexibleResponseStatus(t *testing.T) {
	// Test Case 1: Array of status objects (e.g. from list endpoint)
	listPayload := `[{"status_code": 2000, "status": "success"}]`
	var arrStatus sdp.FlexibleResponseStatus
	if err := json.Unmarshal([]byte(listPayload), &arrStatus); err != nil {
		t.Fatalf("failed to unmarshal array status: %v", err)
	}
	if len(arrStatus) != 1 || arrStatus[0].StatusCode != 2000 {
		t.Fatalf("unexpected arrStatus: %+v", arrStatus)
	}
	if !arrStatus.IsSuccess() {
		t.Fatal("expected IsSuccess to be true")
	}

	// Test Case 2: Single status object (e.g. from get endpoint)
	objPayload := `{"status_code": 2000, "status": "success"}`
	var objStatus sdp.FlexibleResponseStatus
	if err := json.Unmarshal([]byte(objPayload), &objStatus); err != nil {
		t.Fatalf("failed to unmarshal object status: %v", err)
	}
	if len(objStatus) != 1 || objStatus[0].StatusCode != 2000 {
		t.Fatalf("unexpected objStatus: %+v", objStatus)
	}
	if !objStatus.IsSuccess() {
		t.Fatal("expected IsSuccess to be true")
	}
}

func TestUnmarshalSingleRequestWithCustomUDF(t *testing.T) {
	rawJSON := `{
		"request": {
			"id": "606",
			"subject": "Test maintance for multiselect field request",
			"short_description": "",
			"requester": {
				"id": "351",
				"name": "Bayden.Djo",
				"email_id": "Bayden.Djo@Agneko.Forest",
				"mobile": "+7(999)6666666",
				"site": {
					"id": "1501",
					"name": "Noviya Site"
				},
				"department": {
					"id": "1501",
					"name": "TestDep1",
					"deleted": false
				}
			},
			"udf_fields": {
				"udf_multiselect_8191": [
					{
						"name": "marakushev@agneko.com",
						"id": "2106"
					},
					{
						"name": "gusev@agneko.com",
						"id": "2107"
					}
				]
			},
			"status": {
				"id": "2",
				"name": "Open",
				"color": "#0066ff"
			},
			"created_time": {
				"display_value": "02/11/2026 12:00 PM",
				"value": "1793610000166"
			}
		},
		"response_status": {
			"status_code": 2000,
			"status": "success"
		}
	}`

	type MyUDF struct {
		Approvers []sdp.Ref `json:"udf_multiselect_8191"`
	}

	var resp sdp.SingleRequestResponse[MyUDF]
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		t.Fatalf("failed to unmarshal SingleRequestResponse: %v", err)
	}

	req := resp.Request
	if req.ID != "606" {
		t.Errorf("expected ID '606', got %q", req.ID)
	}
	if req.Subject != "Test maintance for multiselect field request" {
		t.Errorf("unexpected subject: %q", req.Subject)
	}
	if req.Status == nil || req.Status.Name != "Open" {
		t.Errorf("unexpected status: %+v", req.Status)
	}
	if req.Requester == nil || req.Requester.Name != "Bayden.Djo" {
		t.Errorf("unexpected requester: %+v", req.Requester)
	}

	// Verify typed UDF custom fields:
	if len(req.UDFFields.Approvers) != 2 {
		t.Fatalf("expected 2 approvers in UDF, got %d", len(req.UDFFields.Approvers))
	}
	if req.UDFFields.Approvers[0].Name != "marakushev@agneko.com" || req.UDFFields.Approvers[0].ID != "2106" {
		t.Errorf("unexpected first approver: %+v", req.UDFFields.Approvers[0])
	}
	if req.UDFFields.Approvers[1].Name != "gusev@agneko.com" || req.UDFFields.Approvers[1].ID != "2107" {
		t.Errorf("unexpected second approver: %+v", req.UDFFields.Approvers[1])
	}
}

func TestUnmarshalListRequests(t *testing.T) {
	rawJSON := `{
		"response_status": [
			{
				"status_code": 2000,
				"status": "success"
			}
		],
		"list_info": {
			"has_more_rows": true,
			"start_index": 1,
			"page": 1,
			"row_count": 2
		},
		"requests": [
			{
				"id": "607",
				"subject": "Ticket 607",
				"status": {
					"id": "2",
					"name": "Open",
					"color": "#0066ff"
				}
			},
			{
				"id": "606",
				"subject": "Ticket 606",
				"status": {
					"id": "2",
					"name": "Open",
					"color": "#0066ff"
				}
			}
		]
	}`

	var resp sdp.RequestListResponse[map[string]any]
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		t.Fatalf("failed to unmarshal RequestListResponse: %v", err)
	}

	if len(resp.Requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(resp.Requests))
	}
	if resp.ListInfo.Page != 1 || resp.ListInfo.RowCount != 2 || !resp.ListInfo.HasMoreRows {
		t.Errorf("unexpected ListInfo: %+v", resp.ListInfo)
	}
	if resp.Requests[0].ID != "607" || resp.Requests[1].ID != "606" {
		t.Errorf("unexpected IDs: %s, %s", resp.Requests[0].ID, resp.Requests[1].ID)
	}
}

func TestAttachmentSizeUnmarshal(t *testing.T) {
	rawJSON := `{
		"request": {
			"id": "700",
			"subject": "Ticket with attachments",
			"attachments": [
				{
					"id": "1",
					"name": "report.pdf",
					"size": {"display_value": "25.4 KB", "value": 26012}
				},
				{
					"id": "2",
					"name": "notes.txt",
					"size": {"display_value": "1.2 KB", "value": "1228"}
				},
				{
					"id": "3",
					"name": "legacy_str.log",
					"size": "512 bytes"
				},
				{
					"id": "4",
					"name": "legacy_num.bin",
					"size": 4096
				},
				{
					"id": "5",
					"name": "no_size.txt"
				}
			]
		},
		"response_status": {"status_code": 2000, "status": "success"}
	}`

	var resp sdp.SingleRequestResponse[map[string]any]
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		t.Fatalf("failed to unmarshal SingleRequestResponse with attachments: %v", err)
	}

	atts := resp.Request.Attachments
	if len(atts) != 5 {
		t.Fatalf("expected 5 attachments, got %d", len(atts))
	}

	if atts[0].Size == nil || atts[0].Size.DisplayValue != "25.4 KB" || atts[0].Size.Value != "26012" || atts[0].Size.String() != "25.4 KB" {
		t.Errorf("unexpected attachment[0] size: %+v", atts[0].Size)
	}
	if atts[1].Size == nil || atts[1].Size.DisplayValue != "1.2 KB" || atts[1].Size.Value != "1228" {
		t.Errorf("unexpected attachment[1] size: %+v", atts[1].Size)
	}
	if atts[2].Size == nil || atts[2].Size.DisplayValue != "" || atts[2].Size.Value != "512 bytes" || atts[2].Size.String() != "512 bytes" {
		t.Errorf("unexpected attachment[2] size: %+v", atts[2].Size)
	}
	if atts[3].Size == nil || atts[3].Size.DisplayValue != "" || atts[3].Size.Value != "4096" || atts[3].Size.String() != "4096" {
		t.Errorf("unexpected attachment[3] size: %+v", atts[3].Size)
	}
	if atts[4].Size != nil {
		t.Errorf("expected nil size for attachment[4], got %+v", atts[4].Size)
	}
}
