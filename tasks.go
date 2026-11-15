package sdp

import (
	"context"
	"fmt"
	"net/url"
)

// StandardTask contains the common, documented fields found in SDP request tasks.
type StandardTask struct {
	ID                 string       `json:"id,omitempty"`
	Title              string       `json:"title,omitempty"`
	Description        *string      `json:"description,omitempty"`
	Status             *StatusRef   `json:"status,omitempty"`
	Owner              *UserRef     `json:"owner,omitempty"`
	Group              *Ref         `json:"group,omitempty"`
	CreatedBy          *UserRef     `json:"created_by,omitempty"`
	CreatedTime        *SDPTime     `json:"created_time,omitempty"`
	ScheduledStartTime *SDPTime     `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   *SDPTime     `json:"scheduled_end_time,omitempty"`
	ActualStartTime    *SDPTime     `json:"actual_start_time,omitempty"`
	ActualEndTime      *SDPTime     `json:"actual_end_time,omitempty"`
	PercentComplete    int          `json:"percent_complete,omitempty"`
	Priority           *Ref         `json:"priority,omitempty"`
	TaskType           *Ref         `json:"task_type,omitempty"`
	Template           *Ref         `json:"template,omitempty"`
	Request            *Ref         `json:"request,omitempty"`
	Site               *Ref         `json:"site,omitempty"`
	IsCancelled        bool         `json:"is_cancelled,omitempty"`
	MarkedOwner        *UserRef     `json:"marked_owner,omitempty"`
	MarkedGroup        *Ref         `json:"marked_group,omitempty"`
	AdditionalCost     string       `json:"additional_cost,omitempty"`
	Attachments        []Attachment `json:"attachments,omitempty"`
}

// Task embeds StandardTask and supports strongly-typed or dynamic custom fields via UDF.
type Task[UDF any] struct {
	StandardTask
	UDFFields UDF `json:"udf_fields,omitempty"`
}

// DefaultTask is an alias for tasks with dynamic map[string]any custom fields.
type DefaultTask = Task[map[string]any]

// SingleTaskResponse wraps GET/POST/PUT /api/v3/requests/{request_id}/tasks/{id}.
type SingleTaskResponse[UDF any] struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	Task           Task[UDF]              `json:"task"`
}

// TaskListResponse wraps GET /api/v3/requests/{request_id}/tasks.
type TaskListResponse[UDF any] struct {
	ResponseStatus FlexibleResponseStatus `json:"response_status"`
	ListInfo       ListInfo               `json:"list_info"`
	Tasks          []Task[UDF]            `json:"tasks"`
}

// ListTasks fetches tasks associated with a request using default dynamic custom fields.
func (s *RequestsService) ListTasks(ctx context.Context, requestID string, opts *ListOptions) (*TaskListResponse[map[string]any], error) {
	return ListRequestTasks[map[string]any](ctx, s.client, requestID, opts)
}

// GetTask fetches a single task associated with a request using default dynamic custom fields.
func (s *RequestsService) GetTask(ctx context.Context, requestID, taskID string) (*DefaultTask, error) {
	return GetRequestTask[map[string]any](ctx, s.client, requestID, taskID)
}

// CreateTask creates a new task under a request using default dynamic custom fields.
func (s *RequestsService) CreateTask(ctx context.Context, requestID string, task any) (*DefaultTask, error) {
	return CreateRequestTask[map[string]any](ctx, s.client, requestID, task)
}

// UpdateTask updates an existing task under a request using default dynamic custom fields.
func (s *RequestsService) UpdateTask(ctx context.Context, requestID, taskID string, task any) (*DefaultTask, error) {
	return UpdateRequestTask[map[string]any](ctx, s.client, requestID, taskID, task)
}

// DeleteTask deletes a task associated with a request.
func (s *RequestsService) DeleteTask(ctx context.Context, requestID, taskID string) error {
	path := fmt.Sprintf("/api/v3/requests/%s/tasks/%s", url.PathEscape(requestID), url.PathEscape(taskID))
	return s.client.delete(ctx, path)
}

// --- Generic Package-Level Functions for Typed Task UDFs ---

// ListRequestTasks fetches tasks for a request matching ListOptions with typed custom fields.
func ListRequestTasks[UDF any](ctx context.Context, client *Client, requestID string, opts *ListOptions) (*TaskListResponse[UDF], error) {
	queryParams, err := encodeListOptions(opts)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v3/requests/%s/tasks", url.PathEscape(requestID))
	var resp TaskListResponse[UDF]
	if err := client.get(ctx, path, queryParams, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetRequestTask fetches a single task for a request with typed custom fields.
func GetRequestTask[UDF any](ctx context.Context, client *Client, requestID, taskID string) (*Task[UDF], error) {
	path := fmt.Sprintf("/api/v3/requests/%s/tasks/%s", url.PathEscape(requestID), url.PathEscape(taskID))
	var resp SingleTaskResponse[UDF]
	if err := client.get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Task, nil
}

// CreateRequestTask creates a new task under a request with typed custom fields.
func CreateRequestTask[UDF any](ctx context.Context, client *Client, requestID string, task any) (*Task[UDF], error) {
	path := fmt.Sprintf("/api/v3/requests/%s/tasks", url.PathEscape(requestID))
	payload := map[string]any{
		"task": task,
	}

	var resp SingleTaskResponse[UDF]
	if err := client.post(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Task, nil
}

// UpdateRequestTask updates an existing task under a request with typed custom fields.
func UpdateRequestTask[UDF any](ctx context.Context, client *Client, requestID, taskID string, task any) (*Task[UDF], error) {
	path := fmt.Sprintf("/api/v3/requests/%s/tasks/%s", url.PathEscape(requestID), url.PathEscape(taskID))
	payload := map[string]any{
		"task": task,
	}

	var resp SingleTaskResponse[UDF]
	if err := client.put(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	return &resp.Task, nil
}
