package report

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestGeneratePDF(t *testing.T) {
	t.Parallel()

	doc := Document{
		Title:            "Service SLA Report",
		Period:           "30d",
		GeneratedAt:      time.Date(2026, time.August, 18, 10, 0, 0, 0, time.UTC),
		UptimeSLATarget:  99.95,
		LatencySLATarget: 99.80,
		FooterText:       "Confidential - SLA Monitoring",
		Endpoints: []EndpointStatistics{
			{
				Name:           "WebSocket Price Stream",
				Group:          "Europe - Frankfurt",
				Type:           "WEBSOCKET",
				URL:            "ws://ws.price.usenobi.com/v1",
				UptimePercent:  95.74,
				LatencyPercent: 98.30,
			},
			{
				Name:           "Auth Service",
				Group:          "Backend Services",
				URL:            "https://auth.usenobi.com/health",
				UptimePercent:  99.50,
				LatencyPercent: 98.80,
			},
		},
	}
	pdf, err := GeneratePDF(doc)
	if err != nil {
		t.Fatalf("unexpected error generating PDF: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatal("expected output to have %PDF header")
	}
}

func TestGeneratePDF_EmptyEndpoints(t *testing.T) {
	t.Parallel()

	doc := Document{
		Title:       "Empty Service SLA Report",
		Period:      "7d",
		GeneratedAt: time.Now(),
		SLATarget:   99.5,
		Endpoints:   nil,
	}
	pdf, err := GeneratePDF(doc)
	if err != nil {
		t.Fatalf("unexpected error generating PDF: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatal("expected output to have %PDF header")
	}
}

func TestGeneratePDF_WithLogo(t *testing.T) {
	t.Parallel()

	// 1x1 base64 transparent PNG
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	tmpFile := t.TempDir() + "/logo.png"
	if err := os.WriteFile(tmpFile, pngData, 0644); err != nil {
		t.Fatal(err)
	}

	doc := Document{
		Title:            "Nobi Labs",
		Logo:             tmpFile,
		Period:           "30d",
		GeneratedAt:      time.Now(),
		UptimeSLATarget:  99.9,
		LatencySLATarget: 99.9,
		Endpoints: []EndpointStatistics{
			{
				Name:           "Price Provider",
				Group:          "Europe - Frankfurt",
				URL:            "ws://ws.price.usenobi.com/v1",
				UptimePercent:  99.95,
				LatencyPercent: 99.85,
			},
		},
	}
	pdf, err := GeneratePDF(doc)
	if err != nil {
		t.Fatalf("unexpected error generating PDF with logo: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatal("expected output to have %PDF header")
	}
}

func TestGeneratePDF_MultiPagePagination(t *testing.T) {
	t.Parallel()

	var endpoints []EndpointStatistics
	for i := 1; i <= 30; i++ {
		endpoints = append(endpoints, EndpointStatistics{
			Name:           fmt.Sprintf("Microservice-%02d", i),
			Group:          fmt.Sprintf("Region-%d", (i%3)+1),
			UptimePercent:  99.0 + float64(i%10)/10.0,
			LatencyPercent: 98.5 + float64(i%15)/10.0,
		})
	}

	doc := Document{
		Title:       "Enterprise Multi-Service SLA Report",
		Period:      "1d",
		GeneratedAt: time.Now(),
		SLATarget:   99.9,
		Endpoints:   endpoints,
	}
	pdf, err := GeneratePDF(doc)
	if err != nil {
		t.Fatalf("unexpected error generating multi-page PDF: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatal("expected output to have %PDF header")
	}
}

func TestFormatEndpointType(t *testing.T) {
	t.Parallel()

	for _, scenario := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "websocket", input: "WEBSOCKET", want: "Websocket"},
		{name: "empty", input: "", want: "Unknown"},
		{name: "unknown", input: "UNKNOWN", want: "Unknown"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if got := formatEndpointType(scenario.input); got != scenario.want {
				t.Errorf("formatEndpointType(%q) = %q, want %q", scenario.input, got, scenario.want)
			}
		})
	}
}

func TestPeriodLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"30d", "Last 30 days"},
		{"7d", "Last 7 days"},
		{"1d", "Last 24 hours"},
		{"1h", "Last hour"},
		{"custom", "custom"},
		{"", "Default period"},
	}

	for _, tt := range tests {
		got := periodLabel(tt.input)
		if got != tt.expected {
			t.Errorf("periodLabel(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}
