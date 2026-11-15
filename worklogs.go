package sdp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// TimeSpent represents the duration logged in a worklog (hours, minutes, and total ms value).
type TimeSpent struct {
	Hours   string `json:"hours,omitempty"`
	Minutes string `json:"minutes,omitempty"`
	Value   string `json:"value,omitempty"` // Total duration in milliseconds as string
}

// UnmarshalJSON decodes TimeSpent while tolerating both string and numeric values.
func (ts *TimeSpent) UnmarshalJSON(data []byte) error {
	var raw struct {
		Hours   json.RawMessage `json:"hours"`
		Minutes json.RawMessage `json:"minutes"`
		Value   json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	ts.Hours = unmarshalFlexibleString(raw.Hours)
	ts.Minutes = unmarshalFlexibleString(raw.Minutes)
	ts.Value = unmarshalFlexibleString(raw.Value)
	return nil
}

// Worklog represents a time entry / worklog associated with an SDP request.
type Worklog struct {
	ID                       string       `json:"id,omitempty"`
	Description              *string      `json:"description,omitempty"`
	Owner                    *UserRef     `json:"owner,omitempty"`
	Technician               *UserRef     `json:"technician,omitempty"`
	TimeSpent                *TimeSpent   `json:"time_spent,omitempty"`
	StartTime                *SDPTime     `json:"start_time,omitempty"`
	EndTime                  *SDPTime     `json:"end_time,omitempty"`
	CreatedBy                *UserRef     `json:"created_by,omitempty"`
	CreatedTime              *SDPTime     `json:"created_time,omitempty"`
	LastUpdatedTime          *SDPTime     `json:"last_updated_time,omitempty"`
	RecordedTime             *SDPTime     `json:"recorded_time,omitempty"`
	Type                     *Ref         `json:"type,omitempty"`
	WorklogType              *Ref         `json:"worklog_type,omitempty"`
	Request                  *Ref         `json:"request,omitempty"`
	Task                     *Ref         `json:"task,omitempty"`
	ChargeOutRate            float64      `json:"charge_out_rate,omitempty"`
	OtherCharge              string       `json:"other_charge,omitempty"`
	TotalCharge              string       `json:"total_charge,omitempty"`
	IsChargesBilled          bool         `json:"is_charges_billed,omitempty"`
	IncludeNonoperationalHrs bool         `json:"include_nonoperational_hours,omitempty"`
	MarkFirstResponse        bool         `json:"mark_first_response,omitempty"`
	Attachments              []Attachment `json:"attachments,omitempty"`
}

// SingleWorklogResponse wraps GET/POST/PUT /api/v3/requests/{request_id}/worklogs/{id}.
type SingleWorklogResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	Worklog        Worklog                `json:"worklog"`
}

// WorklogListResponse wraps GET /api/v3/requests/{request_id}/worklogs.
type WorklogListResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	ListInfo       ListInfo               `json:"list_info"`
	Worklogs       []Worklog              `json:"worklogs"`
}

// ListWorklogs fetches time entries (worklogs) associated with a request.
func (s *RequestsService) ListWorklogs(ctx context.Context, requestID string, opts *ListOptions) (*WorklogListResponse, error) {
	queryParams, err := encodeListOptions(opts)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v3/requests/%s/worklogs", url.PathEscape(requestID))
	var resp WorklogListResponse
	if err := s.client.get(ctx, path, queryParams, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetWorklog fetches a single time entry (worklog) associated with a request.
func (s *RequestsService) GetWorklog(ctx context.Context, requestID, worklogID string) (*Worklog, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/worklogs/%s", url.PathEscape(requestID), url.PathEscape(worklogID))
	var resp SingleWorklogResponse
	if err := s.client.get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Worklog, nil
}

// CreateWorklog adds a new time entry (worklog) to a request.
func (s *RequestsService) CreateWorklog(ctx context.Context, requestID string, worklog any) (*Worklog, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/worklogs", url.PathEscape(requestID))
	payload := map[string]any{
		"worklog": worklog,
	}

	var resp SingleWorklogResponse
	if err := s.client.post(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Worklog, nil
}

// UpdateWorklog updates an existing time entry (worklog) on a request.
func (s *RequestsService) UpdateWorklog(ctx context.Context, requestID, worklogID string, worklog any) (*Worklog, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/worklogs/%s", url.PathEscape(requestID), url.PathEscape(worklogID))
	payload := map[string]any{
		"worklog": worklog,
	}

	var resp SingleWorklogResponse
	if err := s.client.put(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Worklog, nil
}

// DeleteWorklog deletes a time entry (worklog) from a request.
func (s *RequestsService) DeleteWorklog(ctx context.Context, requestID, worklogID string) error {
	path := fmt.Sprintf("/api/v3/requests/%s/worklogs/%s", url.PathEscape(requestID), url.PathEscape(worklogID))
	return s.client.delete(ctx, path)
}
