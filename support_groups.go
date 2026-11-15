package sdp

import (
	"context"
	"fmt"
	"net/url"
)

// SupportGroup represents a support group returned by SDP /api/v3/support_groups endpoints.
type SupportGroup struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Deleted     bool   `json:"deleted,omitempty"`
	Technicians []Ref  `json:"technicians,omitempty"`
}

// SingleSupportGroupResponse wraps GET /api/v3/support_groups/{id}.
type SingleSupportGroupResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	SupportGroup   SupportGroup           `json:"support_group"`
}

// SupportGroupListResponse wraps GET /api/v3/support_groups.
type SupportGroupListResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	ListInfo       ListInfo               `json:"list_info"`
	SupportGroups  []SupportGroup         `json:"support_groups"`
}

// SupportGroupsService handles communication with the support_groups-related methods of the SDP API.
type SupportGroupsService struct {
	client *Client
}

// NewSupportGroupsService creates a new SupportGroupsService.
func NewSupportGroupsService(client *Client) *SupportGroupsService {
	return &SupportGroupsService{client: client}
}

// List fetches support groups matching ListOptions.
func (s *SupportGroupsService) List(ctx context.Context, opts *ListOptions) (*SupportGroupListResponse, error) {
	queryParams, err := encodeListOptions(opts)
	if err != nil {
		return nil, err
	}

	var resp SupportGroupListResponse
	if err := s.client.get(ctx, "/api/v3/support_groups", queryParams, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Get fetches a single support group by ID, including its member technicians.
func (s *SupportGroupsService) Get(ctx context.Context, id string) (*SupportGroup, error) {
	path := fmt.Sprintf("/api/v3/support_groups/%s", url.PathEscape(id))
	var resp SingleSupportGroupResponse
	if err := s.client.get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.SupportGroup, nil
}
