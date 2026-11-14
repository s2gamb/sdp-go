package sdp

import (
	"fmt"
	"strings"
)

// APIError represents an error returned by the ServiceDesk Plus API.
type APIError struct {
	HTTPStatusCode int
	ResponseStatus FlexibleResponseStatus
	RawBody        string
}

func (e *APIError) Error() string {
	var msgs []string
	for _, status := range e.ResponseStatus {
		statusStr := fmt.Sprintf("code=%d status=%s", status.StatusCode, status.Status)
		for _, m := range status.Messages {
			if m.Message != "" {
				if m.Field != "" {
					msgs = append(msgs, fmt.Sprintf("%s (%s: %s)", statusStr, m.Field, m.Message))
				} else {
					msgs = append(msgs, fmt.Sprintf("%s (%s)", statusStr, m.Message))
				}
			}
		}
		if len(status.Messages) == 0 {
			msgs = append(msgs, statusStr)
		}
	}

	if len(msgs) > 0 {
		return fmt.Sprintf("sdp: api error (HTTP %d): %s", e.HTTPStatusCode, strings.Join(msgs, "; "))
	}

	if e.RawBody != "" {
		return fmt.Sprintf("sdp: api error (HTTP %d): %s", e.HTTPStatusCode, e.RawBody)
	}

	return fmt.Sprintf("sdp: api error (HTTP %d)", e.HTTPStatusCode)
}
