package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TwiN/gatus/v5/config"
	"github.com/TwiN/gatus/v5/config/endpoint"
	"github.com/TwiN/gatus/v5/config/maintenance"
	"github.com/TwiN/gatus/v5/config/ui"
	"github.com/TwiN/gatus/v5/storage"
	"github.com/TwiN/gatus/v5/storage/store"
)

func TestSingleEndpointMode(t *testing.T) {
	defer store.Get().Clear()
	defer cache.Clear()
	selected := &endpoint.ExternalEndpoint{Name: "WebSocket", Group: "Europe", Token: "selected-token"}
	hidden := &endpoint.ExternalEndpoint{Name: "API", Group: "Europe", Token: "hidden-token"}
	cfg := &config.Config{
		ExternalEndpoints: []*endpoint.ExternalEndpoint{selected, hidden},
		Maintenance:       &maintenance.Config{},
		Storage: &storage.Config{
			MaximumNumberOfResults: storage.DefaultMaximumNumberOfResults,
			MaximumNumberOfEvents:  storage.DefaultMaximumNumberOfEvents,
		},
		UI: &ui.Config{SingleEndpoint: selected.Key()},
	}
	now := time.Now()
	if err := store.Get().InsertEndpointResult(selected.ToEndpoint(), &endpoint.Result{Success: true, Connected: true, Duration: time.Millisecond, Timestamp: now}); err != nil {
		t.Fatal("failed to insert selected endpoint result:", err)
	}
	if err := store.Get().InsertEndpointResult(hidden.ToEndpoint(), &endpoint.Result{Success: false, Connected: true, Duration: time.Second, Timestamp: now}); err != nil {
		t.Fatal("failed to insert hidden endpoint result:", err)
	}
	router := New(cfg).Router()

	t.Run("root-redirects-to-selected-endpoint", func(t *testing.T) {
		response, err := router.Test(httptest.NewRequest(http.MethodGet, "/", http.NoBody))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusFound {
			t.Fatalf("expected status %d, got %d", http.StatusFound, response.StatusCode)
		}
		if location := response.Header.Get("Location"); location != "/endpoints/europe_websocket" {
			t.Errorf("expected redirect location /endpoints/europe_websocket, got %q", location)
		}
	})

	t.Run("selected-endpoint-page-remains-accessible", func(t *testing.T) {
		response, err := router.Test(httptest.NewRequest(http.MethodGet, "/endpoints/europe_websocket", http.NoBody))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `singleEndpoint: "europe_websocket"`) {
			t.Error("expected selected endpoint key to be injected into the frontend configuration")
		}
	})

	hiddenViewerPaths := []string{
		"/endpoints/europe_api",
		"/api/v1/endpoints/europe_api/statuses",
		"/api/v1/endpoints/europe_api/health/badge.svg",
		"/api/v1/endpoints/europe_api/health/badge.shields",
		"/api/v1/endpoints/europe_api/uptimes/1h",
		"/api/v1/endpoints/europe_api/uptimes/1h/badge.svg",
		"/api/v1/endpoints/europe_api/response-times/1h",
		"/api/v1/endpoints/europe_api/response-times/1h/badge.svg",
		"/api/v1/endpoints/europe_api/response-times/24h/chart.svg",
		"/api/v1/endpoints/europe_api/response-times/24h/history",
	}
	for _, path := range hiddenViewerPaths {
		t.Run("hidden-viewer-route-"+path, func(t *testing.T) {
			response, err := router.Test(httptest.NewRequest(http.MethodGet, path, http.NoBody))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusNotFound {
				t.Errorf("GET %s: expected status %d, got %d", path, http.StatusNotFound, response.StatusCode)
			}
		})
	}

	t.Run("endpoint-list-only-exposes-selected-endpoint", func(t *testing.T) {
		response, err := router.Test(httptest.NewRequest(http.MethodGet, "/api/v1/endpoints/statuses", http.NoBody))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		var statuses []*endpoint.Status
		if err = json.NewDecoder(response.Body).Decode(&statuses); err != nil {
			t.Fatal("failed to decode endpoint statuses:", err)
		}
		if len(statuses) != 1 || statuses[0].Key != selected.Key() {
			t.Errorf("expected only endpoint %q, got %+v", selected.Key(), statuses)
		}
	})

	t.Run("hidden-external-endpoint-ingestion-remains-operational", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/endpoints/europe_api/external?success=true", http.NoBody)
		request.Header.Set("Authorization", "Bearer hidden-token")
		response, err := router.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
	})
}
