package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/TwiN/gatus/v5/config"
	"github.com/TwiN/gatus/v5/report"
	"github.com/TwiN/gatus/v5/storage/store"
	"github.com/gofiber/fiber/v2"
)

var supportedDateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseTimeValue(val string, loc *time.Location, isEnd bool) (time.Time, error) {
	val = strings.TrimSpace(val)
	for _, layout := range supportedDateLayouts {
		if t, err := time.ParseInLocation(layout, val, loc); err == nil {
			if layout == "2006-01-02" && isEnd {
				t = t.Add(24*time.Hour - time.Nanosecond)
			}
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %q (expected YYYY-MM-DD or RFC3339)", val)
}

func parseReportRange(c *fiber.Ctx, cfg *config.Config, now time.Time) (time.Time, time.Time, string, string, error) {
	loc, err := time.LoadLocation(cfg.Reporting.PDF.Timezone)
	if err != nil {
		loc = time.UTC
	}

	fromQuery := c.Query("from", c.Query("start"))
	toQuery := c.Query("to", c.Query("end"))

	if fromQuery != "" || toQuery != "" {
		if fromQuery == "" || toQuery == "" {
			return time.Time{}, time.Time{}, "", "", fmt.Errorf("both start/from and end/to dates must be specified")
		}
		startTime, err := parseTimeValue(fromQuery, loc, false)
		if err != nil {
			return time.Time{}, time.Time{}, "", "", err
		}
		endTime, err := parseTimeValue(toQuery, loc, true)
		if err != nil {
			return time.Time{}, time.Time{}, "", "", err
		}
		if endTime.Before(startTime) || endTime.Equal(startTime) {
			return time.Time{}, time.Time{}, "", "", fmt.Errorf("end date must be after start date")
		}
		if endTime.Sub(startTime) > 90*24*time.Hour+time.Minute {
			return time.Time{}, time.Time{}, "", "", fmt.Errorf("selected date range cannot exceed 90 days")
		}
		periodDisplay := fmt.Sprintf("%s – %s", startTime.In(loc).Format("02 Jan 2006"), endTime.In(loc).Format("02 Jan 2006"))
		fileTag := fmt.Sprintf("%s-to-%s", startTime.In(loc).Format("2006-01-02"), endTime.In(loc).Format("2006-01-02"))
		return startTime, endTime, periodDisplay, fileTag, nil
	}

	period := c.Query("period", cfg.Reporting.PDF.DefaultPeriod)

	// Check if period is encoded as a date range (e.g. 2026-06-01..2026-08-18 or 2026-06-01_2026-08-18)
	var splitDelim string
	for _, delim := range []string{"..", "_", ":", "/"} {
		if strings.Contains(period, delim) {
			splitDelim = delim
			break
		}
	}

	if splitDelim != "" {
		parts := strings.Split(period, splitDelim)
		if len(parts) == 2 {
			startTime, err := parseTimeValue(parts[0], loc, false)
			if err != nil {
				return time.Time{}, time.Time{}, "", "", err
			}
			endTime, err := parseTimeValue(parts[1], loc, true)
			if err != nil {
				return time.Time{}, time.Time{}, "", "", err
			}
			if endTime.Before(startTime) || endTime.Equal(startTime) {
				return time.Time{}, time.Time{}, "", "", fmt.Errorf("end date must be after start date")
			}
			if endTime.Sub(startTime) > 90*24*time.Hour+time.Minute {
				return time.Time{}, time.Time{}, "", "", fmt.Errorf("selected date range cannot exceed 90 days")
			}
			periodDisplay := fmt.Sprintf("%s – %s", startTime.In(loc).Format("02 Jan 2006"), endTime.In(loc).Format("02 Jan 2006"))
			fileTag := fmt.Sprintf("%s-to-%s", startTime.In(loc).Format("2006-01-02"), endTime.In(loc).Format("2006-01-02"))
			return startTime, endTime, periodDisplay, fileTag, nil
		}
	}

	var duration time.Duration
	switch period {
	case "90d":
		duration = 90 * 24 * time.Hour
	case "60d":
		duration = 60 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "1d":
		duration = 24 * time.Hour
	case "1h":
		duration = time.Hour
	default:
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("unsupported reporting period: must be a preset (e.g. 30d, 7d, 1d, 1h, 90d) or date range (max 90 days)")
	}

	if !containsReportPeriod(cfg.Reporting.PDF.AllowedPeriods, period) {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("reporting period is not enabled")
	}

	return now.Add(-duration), now, period, period, nil
}

func SLAReport(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cfg.Reporting == nil || cfg.Reporting.PDF == nil || !cfg.Reporting.PDF.IsEnabled() {
			return c.SendStatus(fiber.StatusNotFound)
		}
		now := time.Now()
		startTime, endTime, periodDisplay, fileTag, err := parseReportRange(c, cfg, now)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		logo := ""
		if cfg.Reporting.PDF.Logo != "" {
			logo = cfg.Reporting.PDF.Logo
		} else if cfg.UI != nil && cfg.UI.Logo != "" {
			logo = cfg.UI.Logo
		}

		doc := report.Document{
			Title:            cfg.Reporting.PDF.Title,
			Logo:             logo,
			Period:           periodDisplay,
			GeneratedAt:      now,
			UptimeSLATarget:  cfg.Reporting.PDF.UptimeSLATargetPercent,
			LatencySLATarget: cfg.Reporting.PDF.LatencySLATargetPercent,
			SLATarget:        cfg.Reporting.PDF.UptimeSLATargetPercent,
			FooterText:       cfg.Reporting.PDF.FooterText,
		}
		add := func(key, group, name, endpointType, targetURL string) error {
			uptime, err := store.Get().GetUptimeByKey(key, startTime, endTime)
			if err != nil {
				return err
			}
			latency, err := store.Get().GetResponseTimeSuccessRateByKey(key, startTime, endTime)
			if err != nil {
				return err
			}
			doc.Endpoints = append(doc.Endpoints, report.EndpointStatistics{
				Name:           name,
				Group:          group,
				Type:           endpointType,
				URL:            targetURL,
				UptimePercent:  uptime * 100,
				LatencyPercent: latency * 100,
			})
			return nil
		}
		for _, ep := range cfg.Endpoints {
			if ep.IsEnabled() {
				if err := add(ep.Key(), ep.Group, ep.Name, string(ep.Type()), ep.URL); err != nil {
					return err
				}
			}
		}
		for _, ep := range cfg.ExternalEndpoints {
			if ep.IsEnabled() {
				if err := add(ep.Key(), ep.Group, ep.Name, string(ep.ToEndpoint().Type()), ep.URL); err != nil {
					return err
				}
			}
		}
		pdf, err := report.GeneratePDF(doc)
		if err != nil {
			return fmt.Errorf("generate SLA report: %w", err)
		}
		c.Set(fiber.HeaderContentType, "application/pdf")
		c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=service-sla-report-%s.pdf", fileTag))
		return c.Send(pdf)
	}
}

func containsReportPeriod(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
