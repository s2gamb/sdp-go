package sdp

import (
	"context"
	"fmt"
	"net/url"
)

// Notification represents a reply or notification message in an SDP request conversation.
type Notification struct {
	ID              string       `json:"id,omitempty"`
	Type            string       `json:"type,omitempty"`
	Subject         string       `json:"subject,omitempty"`
	Description     *string      `json:"description,omitempty"`
	HasAttachments  bool         `json:"has_attachments,omitempty"`
	ShowToRequester bool         `json:"show_to_requester,omitempty"`
	SentTime        *SDPTime     `json:"sent_time,omitempty"`
	Sender          *UserRef     `json:"sender,omitempty"`
	To              []UserRef    `json:"to,omitempty"`
	CC              []UserRef    `json:"cc,omitempty"`
	BCC             []UserRef    `json:"bcc,omitempty"`
	Attachments     []Attachment `json:"attachments,omitempty"`
}

// Conversation represents a single entry in a request's conversation history.
// Depending on the SDP build and endpoint, details may appear directly on Conversation
// or nested under Notification, or be referenced via ContentURL.
type Conversation struct {
	ID              string        `json:"id,omitempty"`
	Type            string        `json:"type,omitempty"`
	Subject         string        `json:"subject,omitempty"`
	Description     *string       `json:"description,omitempty"`
	HasAttachments  bool          `json:"has_attachments,omitempty"`
	ShowToRequester bool          `json:"show_to_requester,omitempty"`
	SentTime        *SDPTime      `json:"sent_time,omitempty"`
	CreatedTime     *SDPTime      `json:"created_time,omitempty"`
	Sender          *UserRef      `json:"sender,omitempty"`
	CreatedBy       *UserRef      `json:"created_by,omitempty"`
	To              []UserRef     `json:"to,omitempty"`
	CC              []UserRef     `json:"cc,omitempty"`
	BCC             []UserRef     `json:"bcc,omitempty"`
	ContentURL      string        `json:"content_url,omitempty"`
	Attachments     []Attachment  `json:"attachments,omitempty"`
	Notification    *Notification `json:"notification,omitempty"`
}

// SingleConversationResponse wraps GET /api/v3/requests/{request_id}/conversations/{id}.
type SingleConversationResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	Conversation   Conversation           `json:"conversation"`
}

// ConversationListResponse wraps GET /api/v3/requests/{request_id}/conversations.
type ConversationListResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	ListInfo       ListInfo               `json:"list_info"`
	Conversations  []Conversation         `json:"conversations"`
}

// SingleNotificationResponse wraps notification payloads from /notifications or conversation content_url.
type SingleNotificationResponse struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	Notification   *Notification          `json:"notification,omitempty"`
	Conversation   *Conversation          `json:"conversation,omitempty"`
}

// ListConversations fetches the conversation history associated with a request.
func (s *RequestsService) ListConversations(ctx context.Context, requestID string, opts *ListOptions) (*ConversationListResponse, error) {
	queryParams, err := encodeListOptions(opts)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v3/requests/%s/conversations", url.PathEscape(requestID))
	var resp ConversationListResponse
	if err := s.client.get(ctx, path, queryParams, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetConversation fetches a single conversation entry for a request by conversation ID.
func (s *RequestsService) GetConversation(ctx context.Context, requestID, conversationID string) (*Conversation, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/conversations/%s", url.PathEscape(requestID), url.PathEscape(conversationID))
	var resp SingleConversationResponse
	if err := s.client.get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Conversation, nil
}

// GetConversationContent fetches the full notification/conversation payload from a
// conversation item's content_url (e.g. "/api/v3/requests/827/notifications/11?input_data=...").
func (s *RequestsService) GetConversationContent(ctx context.Context, contentURL string) (*SingleNotificationResponse, error) {
	if contentURL == "" {
		return nil, fmt.Errorf("sdp: contentURL is required")
	}

	var resp SingleNotificationResponse
	if err := s.client.get(ctx, contentURL, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// SendNotification sends a reply or notification on a request via POST /api/v3/requests/{request_id}/notifications.
func (s *RequestsService) SendNotification(ctx context.Context, requestID string, notification any) (*Notification, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/notifications", url.PathEscape(requestID))
	payload := map[string]any{
		"notification": notification,
	}

	var resp SingleNotificationResponse
	if err := s.client.post(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	if resp.Notification == nil {
		return &Notification{}, nil
	}

	return resp.Notification, nil
}
