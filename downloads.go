package sdp

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
)

// DownloadResponse holds the raw binary content and HTTP metadata of a downloaded file or image.
type DownloadResponse struct {
	Data        []byte
	ContentType string
	Filename    string
}

// Download fetches a binary resource (attachment, inline image, profile pic, or content_url)
// from a relative SDP path or full URL on the configured SDP host.
func (c *Client) Download(ctx context.Context, rawURL string) (*DownloadResponse, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("sdp: download URL is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, rawURL, nil, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read download body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.parseErrorResponse(resp.StatusCode, bodyBytes)
	}

	var filename string
	if disp := resp.Header.Get("Content-Disposition"); disp != "" {
		if _, params, err := mime.ParseMediaType(disp); err == nil {
			filename = params["filename"]
		}
	}

	return &DownloadResponse{
		Data:        bodyBytes,
		ContentType: resp.Header.Get("Content-Type"),
		Filename:    filename,
	}, nil
}

// DownloadAttachment downloads an attachment belonging to a request via
// GET /api/v3/requests/{request_id}/attachments/{attachment_id}/download.
func (s *RequestsService) DownloadAttachment(ctx context.Context, requestID, attachmentID string) (*DownloadResponse, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/attachments/%s/download",
		url.PathEscape(requestID), url.PathEscape(attachmentID))
	return s.client.Download(ctx, path)
}

// DownloadTaskAttachment downloads an attachment belonging to a request task via
// GET /api/v3/requests/{request_id}/tasks/{task_id}/attachments/{attachment_id}/download.
func (s *RequestsService) DownloadTaskAttachment(ctx context.Context, requestID, taskID, attachmentID string) (*DownloadResponse, error) {
	path := fmt.Sprintf("/api/v3/requests/%s/tasks/%s/attachments/%s/download",
		url.PathEscape(requestID), url.PathEscape(taskID), url.PathEscape(attachmentID))
	return s.client.Download(ctx, path)
}

// DownloadImage downloads an inline request image by image ID via
// GET /api/v3/requests/images/{image_id}.
// To download an image by its content-url (e.g. "/inline/WorkOrder/5/abc.jpg"),
// use client.Download(ctx, contentURL).
func (s *RequestsService) DownloadImage(ctx context.Context, imageID string) (*DownloadResponse, error) {
	path := fmt.Sprintf("/api/v3/requests/images/%s", url.PathEscape(imageID))
	return s.client.Download(ctx, path)
}
