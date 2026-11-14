package sdp

import (
	"encoding/json"
	"fmt"
)

// ResponseStatus represents a status item returned by SDP APIs.
type ResponseStatus struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Messages   []ResponseMessage `json:"messages,omitempty"`
}

// ResponseMessage represents error or info messages in ResponseStatus.
type ResponseMessage struct {
	StatusCode int    `json:"status_code,omitempty"`
	Field      string `json:"field,omitempty"`
	Type       string `json:"type,omitempty"`
	Message    string `json:"message,omitempty"`
}

// FlexibleResponseStatus unmarshals both a single status object {"status_code": ...}
// and an array of status objects [{"status_code": ...}], which SDP alternates between.
type FlexibleResponseStatus []ResponseStatus

func (f *FlexibleResponseStatus) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = nil
		return nil
	}

	// Case 1: JSON array e.g. [{"status_code": 2000, ...}]
	if data[0] == '[' {
		var list []ResponseStatus
		if err := json.Unmarshal(data, &list); err != nil {
			return fmt.Errorf("failed to unmarshal status array: %w", err)
		}
		*f = list
		return nil
	}

	// Case 2: Single JSON object e.g. {"status_code": 2000, ...}
	var single ResponseStatus
	if err := json.Unmarshal(data, &single); err != nil {
		return fmt.Errorf("failed to unmarshal single status: %w", err)
	}
	*f = []ResponseStatus{single}
	return nil
}

// IsSuccess returns true if the status indicates success (status_code 2000).
func (f FlexibleResponseStatus) IsSuccess() bool {
	if len(f) == 0 {
		return false
	}
	for _, s := range f {
		if s.StatusCode != 2000 && s.Status != "success" {
			return false
		}
	}
	return true
}

// ListInfo holds pagination metadata returned by list endpoints.
type ListInfo struct {
	HasMoreRows bool `json:"has_more_rows"`
	StartIndex  int  `json:"start_index"`
	Page        int  `json:"page"`
	RowCount    int  `json:"row_count"`
}

// SDPTime captures SDP's timestamp structure where epoch millis is returned as a string.
type SDPTime struct {
	DisplayValue string `json:"display_value"`
	Value        string `json:"value"` // Unix timestamp in milliseconds as string
}

// Ref represents a common {id, name} reference object in SDP.
type Ref struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// StatusRef represents the ticket status along with UI color.
type StatusRef struct {
	Ref
	Color string `json:"color,omitempty"`
}

// TemplateRef represents ticket template metadata.
type TemplateRef struct {
	Ref
	IsServiceTemplate bool `json:"is_service_template"`
}

// ProfilePic represents user profile picture info.
type ProfilePic struct {
	ContentURL string `json:"content-url"`
}

// Department represents department details.
type Department struct {
	Ref
	Deleted bool `json:"deleted"`
	Site    *Ref `json:"site,omitempty"`
}

// UserRef represents requester, technician, or creator details.
type UserRef struct {
	Ref
	EmailID    *string     `json:"email_id,omitempty"`
	Mobile     *string     `json:"mobile,omitempty"`
	Phone      *string     `json:"phone,omitempty"`
	IsVIPUser  bool        `json:"is_vipuser"`
	Site       *Ref        `json:"site,omitempty"`
	Department *Department `json:"department,omitempty"`
	ProfilePic *ProfilePic `json:"profile_pic,omitempty"`
}

// Resolution represents resolution content and attachments.
type Resolution struct {
	Content               *string      `json:"content"`
	ResolutionAttachments []Attachment `json:"resolution_attachments,omitempty"`
}

// Attachment represents an attached file.
type Attachment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size string `json:"size,omitempty"`
}
