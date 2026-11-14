package sdp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// StandardRequest contains the common, documented fields found in SDP requests.
type StandardRequest struct {
	ID                       string       `json:"id,omitempty"`
	Subject                  string       `json:"subject,omitempty"`
	Description              *string      `json:"description,omitempty"`
	ShortDescription         string       `json:"short_description,omitempty"`
	Status                   *StatusRef   `json:"status,omitempty"`
	Template                 *TemplateRef `json:"template,omitempty"`
	Requester                *UserRef     `json:"requester,omitempty"`
	CreatedBy                *UserRef     `json:"created_by,omitempty"`
	Technician               *UserRef     `json:"technician,omitempty"`
	Department               *Department  `json:"department,omitempty"`
	Site                     *Ref         `json:"site,omitempty"`
	Priority                 *Ref         `json:"priority,omitempty"`
	Group                    *Ref         `json:"group,omitempty"`
	Category                 *Ref         `json:"category,omitempty"`
	Subcategory              *Ref         `json:"subcategory,omitempty"`
	Item                     *Ref         `json:"item,omitempty"`
	Maintenance              *Ref         `json:"maintenance,omitempty"`

	// Timestamps
	CreatedTime        *SDPTime `json:"created_time,omitempty"`
	DueByTime          *SDPTime `json:"due_by_time,omitempty"`
	ScheduledStartTime *SDPTime `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   *SDPTime `json:"scheduled_end_time,omitempty"`
	AssignedTime       *SDPTime `json:"assigned_time,omitempty"`
	LastUpdatedTime    *SDPTime `json:"last_updated_time,omitempty"`

	// Booleans
	IsOverdue                bool `json:"is_overdue,omitempty"`
	IsServiceRequest         bool `json:"is_service_request,omitempty"`
	IsTrashed                bool `json:"is_trashed,omitempty"`
	HasNotes                 bool `json:"has_notes,omitempty"`
	HasAttachments           bool `json:"has_attachments,omitempty"`
	HasProblem               bool `json:"has_problem,omitempty"`
	CancelRequested          bool `json:"cancel_requested,omitempty"`
	CancelRequestedIsPending bool `json:"cancel_requested_is_pending,omitempty"`

	// Resolution & Metadata
	Resolution  *Resolution  `json:"resolution,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	EmailTo     []string     `json:"email_to,omitempty"`
	EmailCC     []string     `json:"email_cc,omitempty"`
}

// Request embeds StandardRequest and supports strongly-typed or dynamic custom fields via UDF.
type Request[UDF any] struct {
	StandardRequest
	UDFFields UDF `json:"udf_fields,omitempty"`
}

// DefaultRequest is an alias for requests with dynamic map[string]any custom fields.
type DefaultRequest = Request[map[string]any]

// SingleRequestResponse wraps GET /api/v3/requests/{id}.
type SingleRequestResponse[UDF any] struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	Request        Request[UDF]           `json:"request"`
}

// RequestListResponse wraps GET /api/v3/requests.
type RequestListResponse[UDF any] struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	ListInfo       ListInfo               `json:"list_info"`
	Requests       []Request[UDF]         `json:"requests"`
}

// ListOptions specifies parameters for pagination, sorting, and filtering.
type ListOptions struct {
	Page         int               `json:"page,omitempty"`
	RowCount     int               `json:"row_count,omitempty"`
	StartIndex   int               `json:"start_index,omitempty"`
	SearchFields map[string]any    `json:"search_fields,omitempty"`
	SortFields   map[string]string `json:"sort_fields,omitempty"`
	FilterBy     string            `json:"filter_by,omitempty"`
}

// RequestsService handles communication with the requests-related methods of the SDP API.
type RequestsService struct {
	client *Client
}

// NewRequestsService creates a new RequestsService.
func NewRequestsService(client *Client) *RequestsService {
	return &RequestsService{client: client}
}

// Get fetches a single request by ID using default dynamic custom fields.
func (s *RequestsService) Get(ctx context.Context, id string) (*DefaultRequest, error) {
	return GetRequest[map[string]any](ctx, s.client, id)
}

// List fetches requests using default dynamic custom fields.
func (s *RequestsService) List(ctx context.Context, opts *ListOptions) (*RequestListResponse[map[string]any], error) {
	return ListRequests[map[string]any](ctx, s.client, opts)
}

// Create creates a new request using default dynamic custom fields.
func (s *RequestsService) Create(ctx context.Context, req any) (*DefaultRequest, error) {
	return CreateRequest[map[string]any](ctx, s.client, req)
}

// Update updates an existing request by ID using default dynamic custom fields.
func (s *RequestsService) Update(ctx context.Context, id string, req any) (*DefaultRequest, error) {
	return UpdateRequest[map[string]any](ctx, s.client, id, req)
}

// Delete deletes a request by ID.
func (s *RequestsService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v3/requests/%s", url.PathEscape(id))
	return s.client.delete(ctx, path)
}

// --- Generic Package-Level Functions for Typed UDFs ---

// GetRequest fetches a single request with typed custom fields.
func GetRequest[UDF any](ctx context.Context, client *Client, id string) (*Request[UDF], error) {
	path := fmt.Sprintf("/api/v3/requests/%s", url.PathEscape(id))
	var resp SingleRequestResponse[UDF]

	if err := client.get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Request, nil
}

// ListRequests fetches requests matching ListOptions with typed custom fields.
func ListRequests[UDF any](ctx context.Context, client *Client, opts *ListOptions) (*RequestListResponse[UDF], error) {
	queryParams := url.Values{}

	if opts != nil {
		listInfoMap := map[string]any{}
		if opts.Page > 0 {
			listInfoMap["page"] = opts.Page
		}
		if opts.RowCount > 0 {
			listInfoMap["row_count"] = opts.RowCount
		}
		if opts.StartIndex > 0 {
			listInfoMap["start_index"] = opts.StartIndex
		}
		if len(opts.SearchFields) > 0 {
			listInfoMap["search_fields"] = opts.SearchFields
		}
		if len(opts.SortFields) > 0 {
			listInfoMap["sort_fields"] = opts.SortFields
		}
		if opts.FilterBy != "" {
			listInfoMap["filter_by"] = opts.FilterBy
		}

		if len(listInfoMap) > 0 {
			inputData, err := json.Marshal(map[string]any{"list_info": listInfoMap})
			if err != nil {
				return nil, fmt.Errorf("failed to encode list_info: %w", err)
			}
			queryParams.Set("input_data", string(inputData))
		}
	}

	var resp RequestListResponse[UDF]
	if err := client.get(ctx, "/api/v3/requests", queryParams, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateRequest creates a new request with typed custom fields.
func CreateRequest[UDF any](ctx context.Context, client *Client, req any) (*Request[UDF], error) {
	payload := map[string]any{
		"request": req,
	}

	var resp SingleRequestResponse[UDF]
	if err := client.post(ctx, "/api/v3/requests", payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Request, nil
}

// UpdateRequest updates an existing request with typed custom fields.
func UpdateRequest[UDF any](ctx context.Context, client *Client, id string, req any) (*Request[UDF], error) {
	path := fmt.Sprintf("/api/v3/requests/%s", url.PathEscape(id))
	payload := map[string]any{
		"request": req,
	}

	var resp SingleRequestResponse[UDF]
	if err := client.put(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Request, nil
}
