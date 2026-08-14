package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TwiN/gatus/v5/config"
	"github.com/TwiN/gatus/v5/config/endpoint"
	"github.com/TwiN/gatus/v5/config/ui"
	"github.com/TwiN/gatus/v5/storage/store"
	"github.com/TwiN/gatus/v5/watchdog"
)

func TestSinglePageApplication(t *testing.T) {
	defer store.Get().Clear()
	defer cache.Clear()
	disabled := false
	cfg := &config.Config{
		Metrics: true,
		Endpoints: []*endpoint.Endpoint{
			{
				Name:  "frontend",
				Group: "core",
			},
			{
				Name:  "backend",
				Group: "core",
			},
		},
		UI: &ui.Config{
			Title:                         "example-title",
			Description:                   "example-description",
			UptimeStatistics:              &disabled,
			ResponseTimeStatistics:        &disabled,
			CurrentHealth:                 &disabled,
			ResponseTimeTrend:             &disabled,
			Events:                        &disabled,
			ResponseTimeStatisticsPeriods: []string{"30d", "7d"},
			UptimeStatisticsPeriods:       []string{"30d"},
		},
	}
	watchdog.UpdateEndpointStatus(cfg.Endpoints[0], &endpoint.Result{Success: true, Duration: time.Millisecond, Timestamp: time.Now()})
	watchdog.UpdateEndpointStatus(cfg.Endpoints[1], &endpoint.Result{Success: false, Duration: time.Second, Timestamp: time.Now()})
	api := New(cfg)
	router := api.Router()
	type Scenario struct {
		Name              string
		Path              string
		Gzip              bool
		CookieDarkMode    bool
		UIDarkMode        bool
		ExpectedCode      int
		ExpectedDarkTheme bool
	}
	scenarios := []Scenario{
		{
			Name:              "frontend-home",
			Path:              "/",
			CookieDarkMode:    true,
			UIDarkMode:        false,
			ExpectedDarkTheme: true,
			ExpectedCode:      200,
		},
		{
			Name:              "frontend-endpoint-light",
			Path:              "/endpoints/core_frontend",
			CookieDarkMode:    false,
			UIDarkMode:        false,
			ExpectedDarkTheme: false,
			ExpectedCode:      200,
		},
		{
			Name:              "frontend-endpoint-dark",
			Path:              "/endpoints/core_frontend",
			CookieDarkMode:    false,
			UIDarkMode:        true,
			ExpectedDarkTheme: true,
			ExpectedCode:      200,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			cfg.UI.DarkMode = &scenario.UIDarkMode
			request := httptest.NewRequest("GET", scenario.Path, http.NoBody)
			if scenario.Gzip {
				request.Header.Set("Accept-Encoding", "gzip")
			}
			if scenario.CookieDarkMode {
				request.Header.Set("Cookie", "theme=dark")
			}
			response, err := router.Test(request)
			if err != nil {
				return
			}
			defer response.Body.Close()
			if response.StatusCode != scenario.ExpectedCode {
				t.Errorf("%s %s should have returned %d, but returned %d instead", request.Method, request.URL, scenario.ExpectedCode, response.StatusCode)
			}
			body, _ := io.ReadAll(response.Body)
			strBody := string(body)
			if !strings.Contains(strBody, cfg.UI.Title) {
				t.Errorf("%s %s should have contained the title", request.Method, request.URL)
			}
			for _, metadata := range []string{
				`<meta name="description" content="example-description"`,
				`<meta property="og:title" content="example-title"`,
				`<meta property="og:description" content="example-description"`,
				`<meta name="twitter:title" content="example-title"`,
				`<meta name="twitter:description" content="example-description"`,
			} {
				if !strings.Contains(strBody, metadata) {
					t.Errorf("%s %s should have contained metadata %s", request.Method, request.URL, metadata)
				}
			}
			for _, setting := range []string{`uptimeStatistics: "false"`, `responseTimeStatistics: "false"`, `currentHealth: "false"`, `responseTimeTrend: "false"`, `events: "false"`} {
				if !strings.Contains(strBody, setting) {
					t.Errorf("%s %s should have contained %s", request.Method, request.URL, setting)
				}
			}
			if !strings.Contains(strBody, `responseTimeStatisticsPeriods: "30d,7d"`) {
				t.Errorf("%s %s should have contained configured response time statistics periods", request.Method, request.URL)
			}
			if !strings.Contains(strBody, `uptimeStatisticsPeriods: "30d"`) {
				t.Errorf("%s %s should have contained configured uptime statistics periods", request.Method, request.URL)
			}
			if scenario.ExpectedDarkTheme && !strings.Contains(strBody, "class=\"dark\"") {
				t.Errorf("%s %s should have responded with dark mode headers", request.Method, request.URL)
			}
			if !scenario.ExpectedDarkTheme && strings.Contains(strBody, "class=\"dark\"") {
				t.Errorf("%s %s should not have responded with dark mode headers", request.Method, request.URL)
			}
		})
	}
}
