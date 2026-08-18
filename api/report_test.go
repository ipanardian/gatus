package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TwiN/gatus/v5/config"
	"github.com/TwiN/gatus/v5/config/endpoint"
	"github.com/TwiN/gatus/v5/config/maintenance"
	"github.com/TwiN/gatus/v5/config/reporting"
	"github.com/TwiN/gatus/v5/security"
	"github.com/TwiN/gatus/v5/storage"
	"github.com/TwiN/gatus/v5/storage/store"
)

func TestSLAReport(t *testing.T) {
	defer store.Get().Clear()
	enabled := true
	external := &endpoint.ExternalEndpoint{Name: "WebSocket", Group: "Europe", URL: "ws://ws.price.usenobi.com/v1", Token: "token"}
	cfg := &config.Config{
		Security: &security.Config{Basic: &security.BasicConfig{
			Username:                        "nobi",
			PasswordBcryptHashBase64Encoded: "JDJhJDA4JDFoRnpPY1hnaFl1OC9ISlFsa21VS09wOGlPU1ZOTDlHZG1qeTFvb3dIckRBUnlHUmNIRWlT",
		}},
		ExternalEndpoints: []*endpoint.ExternalEndpoint{external},
		Maintenance:       &maintenance.Config{},
		Reporting: &reporting.Config{PDF: &reporting.PDFConfig{
			Enabled:                 &enabled,
			AllowedPeriods:          []string{"90d", "30d", "7d", "1d", "1h"},
			DefaultPeriod:           "30d",
			UptimeSLATargetPercent:  99.9,
			LatencySLATargetPercent: 99.9,
			FooterText:              "Automated SLA Monitoring",
			SuperAdmin: &security.BasicConfig{
				Username:                        "super-admin",
				PasswordBcryptHashBase64Encoded: "JDJhJDA4JDFoRnpPY1hnaFl1OC9ISlFsa21VS09wOGlPU1ZOTDlHZG1qeTFvb3dIckRBUnlHUmNIRWlT",
			},
		}},
		Storage: &storage.Config{
			MaximumNumberOfResults: storage.DefaultMaximumNumberOfResults,
			MaximumNumberOfEvents:  storage.DefaultMaximumNumberOfEvents,
		},
	}
	if err := store.Get().InsertEndpointResult(external.ToEndpoint(), &endpoint.Result{
		Success:   true,
		Connected: true,
		Duration:  time.Millisecond,
		Timestamp: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	reportRequest := func(path string) *http.Request {
		request := httptest.NewRequest(http.MethodGet, path, http.NoBody)
		request.SetBasicAuth("super-admin", "hunter2")
		return request
	}

	t.Run("valid preset period", func(t *testing.T) {
		response, err := New(cfg).Router().Test(reportRequest("/api/v1/reports/sla.pdf?period=30d"))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		if contentType := response.Header.Get("Content-Type"); contentType != "application/pdf" {
			t.Fatalf("expected application/pdf content type, got %q", contentType)
		}
	})

	t.Run("requires super-admin credentials", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/sla.pdf?period=30d", http.NoBody)
		response, err := New(cfg).Router().Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status %d without super-admin credentials, got %d", http.StatusUnauthorized, response.StatusCode)
		}
		if realm := response.Header.Get("WWW-Authenticate"); realm != `Basic realm="SLA Reports"` {
			t.Fatalf("expected dedicated SLA report authentication realm, got %q", realm)
		}

		request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/sla.pdf?period=30d", http.NoBody)
		request.SetBasicAuth("nobi", "hunter2")
		response, err = New(cfg).Router().Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status %d with global credentials, got %d", http.StatusUnauthorized, response.StatusCode)
		}

		request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/sla.pdf?period=30d", http.NoBody)
		request.SetBasicAuth("super-admin", "hunter2")
		response, err = New(cfg).Router().Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d with super-admin credentials, got %d", http.StatusOK, response.StatusCode)
		}
	})

	t.Run("valid start and end date range within 90 days", func(t *testing.T) {
		response, err := New(cfg).Router().Test(reportRequest("/api/v1/reports/sla.pdf?from=2026-06-01&to=2026-08-18"))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		if disposition := response.Header.Get("Content-Disposition"); !bytes.Contains([]byte(disposition), []byte("2026-06-01-to-2026-08-18")) {
			t.Fatalf("expected disposition to contain date range filename, got %q", disposition)
		}
	})

	t.Run("valid start and end in period param", func(t *testing.T) {
		response, err := New(cfg).Router().Test(reportRequest("/api/v1/reports/sla.pdf?period=2026-07-01..2026-08-18"))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
	})

	t.Run("reject date range exceeding 90 days", func(t *testing.T) {
		response, err := New(cfg).Router().Test(reportRequest("/api/v1/reports/sla.pdf?from=2026-01-01&to=2026-08-18"))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
		}
	})

	t.Run("reject invalid inverted date range", func(t *testing.T) {
		response, err := New(cfg).Router().Test(reportRequest("/api/v1/reports/sla.pdf?from=2026-08-18&to=2026-06-01"))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
		}
	})
}
