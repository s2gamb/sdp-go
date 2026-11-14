package sdp

import (
	"context"
	"fmt"
	"net/url"
)

// StandardTechnician contains the common, documented fields found in SDP technicians.
type StandardTechnician struct {
	ID          string      `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	EmailID     *string     `json:"email_id,omitempty"`
	LoginName   string      `json:"login_name,omitempty"`
	Status      string      `json:"status,omitempty"`
	Mobile      *string     `json:"mobile,omitempty"`
	Phone       *string     `json:"phone,omitempty"`
	IsVIPUser   bool        `json:"is_vipuser,omitempty"`
	Site        *Ref        `json:"site,omitempty"`
	Department  *Department `json:"department,omitempty"`
	ReportingTo *UserRef    `json:"reporting_to,omitempty"`
	ProfilePic  *ProfilePic `json:"profile_pic,omitempty"`
	EmployeeID  string      `json:"employee_id,omitempty"`
	JobTitle    string      `json:"job_title,omitempty"`
	CostPerHour string      `json:"cost_per_hour,omitempty"`
	Description *string     `json:"description,omitempty"`
}

// Technician embeds StandardTechnician and supports strongly-typed or dynamic custom fields via UDF.
type Technician[UDF any] struct {
	StandardTechnician
	UDFFields UDF `json:"udf_fields,omitempty"`
}

// DefaultTechnician is an alias for technicians with dynamic map[string]any custom fields.
type DefaultTechnician = Technician[map[string]any]

// SingleTechnicianResponse wraps GET/PUT /api/v3/technicians/{id}.
type SingleTechnicianResponse[UDF any] struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	Technician     Technician[UDF]        `json:"technician"`
}

// TechnicianListResponse wraps GET /api/v3/technicians.
type TechnicianListResponse[UDF any] struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	ListInfo       ListInfo               `json:"list_info"`
	Technicians    []Technician[UDF]      `json:"technicians"`
}

// TechniciansService handles communication with the technicians-related methods of the SDP API.
type TechniciansService struct {
	client *Client
}

// NewTechniciansService creates a new TechniciansService.
func NewTechniciansService(client *Client) *TechniciansService {
	return &TechniciansService{client: client}
}

// Get fetches a single technician by ID using default dynamic custom fields.
func (s *TechniciansService) Get(ctx context.Context, id string) (*DefaultTechnician, error) {
	return GetTechnician[map[string]any](ctx, s.client, id)
}

// List fetches technicians using default dynamic custom fields.
func (s *TechniciansService) List(ctx context.Context, opts *ListOptions) (*TechnicianListResponse[map[string]any], error) {
	return ListTechnicians[map[string]any](ctx, s.client, opts)
}

// Update updates an existing technician by ID using default dynamic custom fields.
func (s *TechniciansService) Update(ctx context.Context, id string, tech any) (*DefaultTechnician, error) {
	return UpdateTechnician[map[string]any](ctx, s.client, id, tech)
}

// FindByField fetches a single technician whose field exactly matches value.
func (s *TechniciansService) FindByField(ctx context.Context, field string, value any) (*DefaultTechnician, error) {
	return FindTechnicianByField[map[string]any](ctx, s.client, field, value)
}

// --- Generic Package-Level Functions for Typed UDFs ---

// GetTechnician fetches a single technician with typed custom fields.
func GetTechnician[UDF any](ctx context.Context, client *Client, id string) (*Technician[UDF], error) {
	path := fmt.Sprintf("/api/v3/technicians/%s", url.PathEscape(id))
	var resp SingleTechnicianResponse[UDF]

	if err := client.get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Technician, nil
}

// ListTechnicians fetches technicians matching ListOptions with typed custom fields.
func ListTechnicians[UDF any](ctx context.Context, client *Client, opts *ListOptions) (*TechnicianListResponse[UDF], error) {
	queryParams, err := encodeListOptions(opts)
	if err != nil {
		return nil, err
	}

	var resp TechnicianListResponse[UDF]
	if err := client.get(ctx, "/api/v3/technicians", queryParams, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateTechnician updates an existing technician with typed custom fields.
func UpdateTechnician[UDF any](ctx context.Context, client *Client, id string, tech any) (*Technician[UDF], error) {
	path := fmt.Sprintf("/api/v3/technicians/%s", url.PathEscape(id))
	payload := map[string]any{
		"technician": tech,
	}

	var resp SingleTechnicianResponse[UDF]
	if err := client.put(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Technician, nil
}

// FindTechnicianByField fetches a single technician whose field exactly matches value.
// It returns ErrTechnicianNotFound if no technician matches, or ErrMultipleTechniciansFound
// if more than one technician matches.
func FindTechnicianByField[UDF any](ctx context.Context, client *Client, field string, value any) (*Technician[UDF], error) {
	if field == "" {
		return nil, fmt.Errorf("sdp: field name is required")
	}

	resp, err := ListTechnicians[UDF](ctx, client, &ListOptions{
		RowCount: 2,
		SearchCriteria: []SearchCriterion{
			{
				Field:     field,
				Condition: "is",
				Value:     value,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	switch len(resp.Technicians) {
	case 0:
		return nil, ErrTechnicianNotFound
	case 1:
		return &resp.Technicians[0], nil
	default:
		return nil, ErrMultipleTechniciansFound
	}
}
