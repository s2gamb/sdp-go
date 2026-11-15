package sdp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s2gamb/sdp-go"
)

func TestDownloads(t *testing.T) {
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	pdfBytes := []byte("%PDF-1.4\n%test")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("authtoken") != "test-token" {
			t.Errorf("missing or invalid authtoken: %q", r.Header.Get("authtoken"))
		}

		switch r.URL.Path {
		case "/api/v3/requests/606/attachments/901/download":
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", `attachment; filename="report.pdf"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pdfBytes)

		case "/api/v3/requests/606/tasks/5564/attachments/902/download":
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Disposition", `attachment; filename="task.png"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngBytes)

		case "/api/v3/requests/images/77":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngBytes)

		case "/inline/WorkOrder/5/abc.png":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngBytes)

		case "/api/v3/requests/606/attachments/404/download":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{
				"response_status": {
					"status_code": 4004,
					"status": "failed",
					"messages": [{"status_code": 4004, "message": "Attachment not found"}]
				}
			}`))

		default:
			t.Errorf("unexpected GET %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, err := sdp.NewClient(ts.URL, sdp.WithAuthToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	// 1. DownloadAttachment
	att, err := client.Requests.DownloadAttachment(ctx, "606", "901")
	if err != nil {
		t.Fatalf("DownloadAttachment failed: %v", err)
	}
	if !bytes.Equal(att.Data, pdfBytes) {
		t.Errorf("unexpected attachment bytes: %v", att.Data)
	}
	if att.ContentType != "application/pdf" {
		t.Errorf("expected application/pdf, got %s", att.ContentType)
	}
	if att.Filename != "report.pdf" {
		t.Errorf("expected report.pdf, got %q", att.Filename)
	}

	// 2. DownloadTaskAttachment
	taskAtt, err := client.Requests.DownloadTaskAttachment(ctx, "606", "5564", "902")
	if err != nil {
		t.Fatalf("DownloadTaskAttachment failed: %v", err)
	}
	if !bytes.Equal(taskAtt.Data, pngBytes) || taskAtt.Filename != "task.png" {
		t.Errorf("unexpected task attachment: %+v", taskAtt)
	}

	// 3. DownloadImage
	img, err := client.Requests.DownloadImage(ctx, "77")
	if err != nil {
		t.Fatalf("DownloadImage failed: %v", err)
	}
	if !bytes.Equal(img.Data, pngBytes) || img.ContentType != "image/png" {
		t.Errorf("unexpected image response: %+v", img)
	}

	// 4. Generic client.Download for inline image content_url
	inlineImg, err := client.Download(ctx, "/inline/WorkOrder/5/abc.png")
	if err != nil {
		t.Fatalf("client.Download failed: %v", err)
	}
	if !bytes.Equal(inlineImg.Data, pngBytes) {
		t.Errorf("unexpected inline image bytes")
	}

	// 5. Error handling returns *sdp.APIError
	_, err = client.Requests.DownloadAttachment(ctx, "606", "404")
	if err == nil {
		t.Fatal("expected error for missing attachment, got nil")
	}
	apiErr, ok := err.(*sdp.APIError)
	if !ok || apiErr.HTTPStatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 *sdp.APIError, got %T: %v", err, err)
	}
}

func TestDownload_DoesNotLeakAuthToExternalHost(t *testing.T) {
	extServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("authtoken") != "" || r.Header.Get("Authorization") != "" {
			t.Errorf("auth header leaked to external host!")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer extServer.Close()

	client, err := sdp.NewClient("https://sdp.example.com", sdp.WithAuthToken("secret"))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	resp, err := client.Download(context.Background(), extServer.URL+"/file.png")
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if string(resp.Data) != "ok" {
		t.Errorf("unexpected data: %q", resp.Data)
	}
}
