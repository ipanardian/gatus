package reporting

import (
	"errors"
	"testing"
)

func TestConfigValidateAndSetDefaults(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		cfg := &PDFConfig{}
		if err := cfg.ValidateAndSetDefaults(); err != nil {
			t.Fatal(err)
		}
		if cfg.IsEnabled() {
			t.Fatal("expected reporting to be disabled by default")
		}
	})

	t.Run("enabled configuration", func(t *testing.T) {
		cfg := &PDFConfig{
			Enabled:                 boolPointer(true),
			AllowedPeriods:          []string{"30d", "7d", "1d", "1h"},
			DefaultPeriod:           "30d",
			UptimeSLATargetPercent:  99.95,
			LatencySLATargetPercent: 99.80,
			FooterText:              "Custom Company SLA Monitoring",
		}
		if err := cfg.ValidateAndSetDefaults(); err != nil {
			t.Fatal(err)
		}
		if !cfg.IsEnabled() {
			t.Fatal("expected reporting to be enabled")
		}
		if cfg.FooterText != "Custom Company SLA Monitoring" {
			t.Fatalf("expected custom footer text, got %q", cfg.FooterText)
		}
		if cfg.UptimeSLATargetPercent != 99.95 || cfg.LatencySLATargetPercent != 99.80 {
			t.Fatal("expected custom SLA targets to be retained")
		}
	})

	t.Run("backward compatibility with legacy sla-target-percent", func(t *testing.T) {
		cfg := &PDFConfig{
			Enabled:          boolPointer(true),
			SLATargetPercent: 99.5,
		}
		if err := cfg.ValidateAndSetDefaults(); err != nil {
			t.Fatal(err)
		}
		if cfg.UptimeSLATargetPercent != 99.5 || cfg.LatencySLATargetPercent != 99.5 {
			t.Fatalf("expected both uptime and latency target to inherit legacy SLATargetPercent 99.5, got uptime=%.2f, latency=%.2f", cfg.UptimeSLATargetPercent, cfg.LatencySLATargetPercent)
		}
		if cfg.FooterText != defaultFooterText {
			t.Fatalf("expected default footer text, got %q", cfg.FooterText)
		}
	})

	t.Run("rejects invalid period", func(t *testing.T) {
		cfg := &PDFConfig{Enabled: boolPointer(true), AllowedPeriods: []string{"2d"}}
		if !errors.Is(cfg.ValidateAndSetDefaults(), ErrInvalidPeriod) {
			t.Fatal("expected invalid period error")
		}
	})

	t.Run("rejects uptime target outside percentage range", func(t *testing.T) {
		cfg := &PDFConfig{Enabled: boolPointer(true), UptimeSLATargetPercent: 100.1}
		if !errors.Is(cfg.ValidateAndSetDefaults(), ErrInvalidUptimeSLATargetPercent) {
			t.Fatal("expected invalid uptime SLA target error")
		}
	})

	t.Run("rejects latency target outside percentage range", func(t *testing.T) {
		cfg := &PDFConfig{Enabled: boolPointer(true), LatencySLATargetPercent: -1}
		if !errors.Is(cfg.ValidateAndSetDefaults(), ErrInvalidLatencySLATargetPercent) {
			t.Fatal("expected invalid latency SLA target error")
		}
	})
}
