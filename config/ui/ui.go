package ui

import (
	"bytes"
	"errors"
	"html/template"
	"strings"

	"github.com/TwiN/gatus/v5/storage"
	static "github.com/TwiN/gatus/v5/web"
)

const (
	defaultTitle               = "Health Dashboard | Gatus"
	defaultDescription         = "Gatus is an advanced automated status page that lets you monitor your applications and configure alerts to notify you if there's an issue"
	defaultHeader              = "Gatus"
	defaultDashboardHeading    = "Health Dashboard"
	defaultDashboardSubheading = "Monitor the health of your endpoints in real-time"
	defaultLogo                = ""
	defaultLink                = ""
	defaultFavicon             = "/favicon.ico"
	defaultFavicon16           = "/favicon-16x16.png"
	defaultFavicon32           = "/favicon-32x32.png"
	defaultCustomCSS           = ""
	defaultSortBy              = "name"
	defaultFilterBy            = "none"
	defaultLoginSubtitle       = "System Monitoring Dashboard"
)

var (
	defaultDarkMode                 = true
	defaultSectionEnabled           = true
	defaultResponseTimeBadgePeriods = []string{"30d", "7d", "24h", "1h"}

	ErrButtonValidationFailed         = errors.New("invalid button configuration: missing required name or link")
	ErrInvalidDefaultSortBy           = errors.New("invalid default-sort-by value: must be 'name', 'group', or 'health'")
	ErrInvalidDefaultFilterBy         = errors.New("invalid default-filter-by value: must be 'none', 'failing', or 'unstable'")
	ErrInvalidResponseTimeBadgePeriod = errors.New("invalid response-time-badge-periods value: must contain only '30d', '7d', '24h', or '1h' without duplicates")
)

// Config is the configuration for the UI of Gatus
type Config struct {
	Title                    string   `yaml:"title,omitempty"`                       // Title of the page
	Description              string   `yaml:"description,omitempty"`                 // Meta description of the page
	DashboardHeading         string   `yaml:"dashboard-heading,omitempty"`           // Dashboard Title between header and endpoints
	DashboardSubheading      string   `yaml:"dashboard-subheading,omitempty"`        // Dashboard Description between header and endpoints
	Header                   string   `yaml:"header,omitempty"`                      // Header is the text at the top of the page
	Logo                     string   `yaml:"logo,omitempty"`                        // Logo to display on the page
	Link                     string   `yaml:"link,omitempty"`                        // Link to open when clicking on the logo
	Favicon                  Favicon  `yaml:"favicon,omitempty"`                     // Favourite icon to display in web browser tab or address bar
	Buttons                  []Button `yaml:"buttons,omitempty"`                     // Buttons to display below the header
	CustomCSS                string   `yaml:"custom-css,omitempty"`                  // Custom CSS to include in the page
	DarkMode                 *bool    `yaml:"dark-mode,omitempty"`                   // DarkMode is a flag to enable dark mode by default
	DefaultSortBy            string   `yaml:"default-sort-by,omitempty"`             // DefaultSortBy is the default sort option ('name', 'group', 'health')
	DefaultFilterBy          string   `yaml:"default-filter-by,omitempty"`           // DefaultFilterBy is the default filter option ('none', 'failing', 'unstable')
	LoginSubtitle            string   `yaml:"login-subtitle,omitempty"`              // LoginSubtitle is the subtitle displayed on the OIDC login page
	UptimeStatistics         *bool    `yaml:"uptime-statistics,omitempty"`           // UptimeStatistics controls whether uptime statistics are displayed on endpoint details pages
	CurrentHealth            *bool    `yaml:"current-health,omitempty"`              // CurrentHealth controls whether the current health badge is displayed on endpoint details pages
	ResponseTimeTrend        *bool    `yaml:"response-time-trend,omitempty"`         // ResponseTimeTrend controls whether the response time trend is displayed on endpoint details pages
	Events                   *bool    `yaml:"events,omitempty"`                      // Events controls whether events are displayed on endpoint details pages
	ResponseTimeBadgePeriods []string `yaml:"response-time-badge-periods,omitempty"` // ResponseTimeBadgePeriods controls which response time badges are displayed and in which order
	//////////////////////////////////////////////
	// Non-configurable - used for UI rendering //
	//////////////////////////////////////////////
	MaximumNumberOfResults int `yaml:"-"` // MaximumNumberOfResults to display on the page, it's not configurable because we're passing it from the storage config
}

func (cfg *Config) IsDarkMode() bool {
	if cfg.DarkMode != nil {
		return *cfg.DarkMode
	}
	return defaultDarkMode
}

func (cfg *Config) IsUptimeStatisticsEnabled() bool {
	return cfg.UptimeStatistics == nil || *cfg.UptimeStatistics
}

func (cfg *Config) IsCurrentHealthEnabled() bool {
	return cfg.CurrentHealth == nil || *cfg.CurrentHealth
}

func (cfg *Config) IsResponseTimeTrendEnabled() bool {
	return cfg.ResponseTimeTrend == nil || *cfg.ResponseTimeTrend
}

func (cfg *Config) IsEventsEnabled() bool {
	return cfg.Events == nil || *cfg.Events
}

func (cfg *Config) ResponseTimeBadgePeriodsCSV() string {
	return strings.Join(cfg.ResponseTimeBadgePeriods, ",")
}

func boolPointer(value bool) *bool {
	return &value
}

// Button is the configuration for a button on the UI
type Button struct {
	Name string `yaml:"name,omitempty"` // Name is the text to display on the button
	Link string `yaml:"link,omitempty"` // Link to open when the button is clicked.
}

// Validate validates the button configuration
func (btn *Button) Validate() error {
	if len(btn.Name) == 0 || len(btn.Link) == 0 {
		return ErrButtonValidationFailed
	}
	return nil
}

type Favicon struct {
	Default   string `yaml:"default,omitempty"`   // URL or path to default favourite icon.
	Size16x16 string `yaml:"size16x16,omitempty"` // URL or path to favourite icon for 16x16 size.
	Size32x32 string `yaml:"size32x32,omitempty"` // URL or path to favourite icon for 32x32 size.
}

// GetDefaultConfig returns a Config struct with the default values
func GetDefaultConfig() *Config {
	return &Config{
		Title:                    defaultTitle,
		Description:              defaultDescription,
		DashboardHeading:         defaultDashboardHeading,
		DashboardSubheading:      defaultDashboardSubheading,
		Header:                   defaultHeader,
		Logo:                     defaultLogo,
		Link:                     defaultLink,
		CustomCSS:                defaultCustomCSS,
		DarkMode:                 &defaultDarkMode,
		DefaultSortBy:            defaultSortBy,
		DefaultFilterBy:          defaultFilterBy,
		LoginSubtitle:            defaultLoginSubtitle,
		UptimeStatistics:         boolPointer(defaultSectionEnabled),
		CurrentHealth:            boolPointer(defaultSectionEnabled),
		ResponseTimeTrend:        boolPointer(defaultSectionEnabled),
		Events:                   boolPointer(defaultSectionEnabled),
		ResponseTimeBadgePeriods: append([]string(nil), defaultResponseTimeBadgePeriods...),
		MaximumNumberOfResults:   storage.DefaultMaximumNumberOfResults,
		Favicon: Favicon{
			Default:   defaultFavicon,
			Size16x16: defaultFavicon16,
			Size32x32: defaultFavicon32,
		},
	}
}

// ValidateAndSetDefaults validates the UI configuration and sets the default values if necessary.
func (cfg *Config) ValidateAndSetDefaults() error {
	if len(cfg.Title) == 0 {
		cfg.Title = defaultTitle
	}
	if len(cfg.Description) == 0 {
		cfg.Description = defaultDescription
	}
	if len(cfg.DashboardHeading) == 0 {
		cfg.DashboardHeading = defaultDashboardHeading
	}
	if len(cfg.DashboardSubheading) == 0 {
		cfg.DashboardSubheading = defaultDashboardSubheading
	}
	if len(cfg.Header) == 0 {
		cfg.Header = defaultHeader
	}
	if len(cfg.Logo) == 0 {
		cfg.Logo = defaultLogo
	}
	if len(cfg.Link) == 0 {
		cfg.Link = defaultLink
	}
	if len(cfg.CustomCSS) == 0 {
		cfg.CustomCSS = defaultCustomCSS
	}
	if cfg.DarkMode == nil {
		cfg.DarkMode = &defaultDarkMode
	}
	if len(cfg.DefaultSortBy) == 0 {
		cfg.DefaultSortBy = defaultSortBy
	} else if cfg.DefaultSortBy != "name" && cfg.DefaultSortBy != "group" && cfg.DefaultSortBy != "health" {
		return ErrInvalidDefaultSortBy
	}
	if len(cfg.DefaultFilterBy) == 0 {
		cfg.DefaultFilterBy = defaultFilterBy
	} else if cfg.DefaultFilterBy != "none" && cfg.DefaultFilterBy != "failing" && cfg.DefaultFilterBy != "unstable" {
		return ErrInvalidDefaultFilterBy
	}
	if len(cfg.LoginSubtitle) == 0 {
		cfg.LoginSubtitle = defaultLoginSubtitle
	}
	if cfg.UptimeStatistics == nil {
		cfg.UptimeStatistics = boolPointer(defaultSectionEnabled)
	}
	if cfg.CurrentHealth == nil {
		cfg.CurrentHealth = boolPointer(defaultSectionEnabled)
	}
	if cfg.ResponseTimeTrend == nil {
		cfg.ResponseTimeTrend = boolPointer(defaultSectionEnabled)
	}
	if cfg.Events == nil {
		cfg.Events = boolPointer(defaultSectionEnabled)
	}
	if cfg.ResponseTimeBadgePeriods == nil {
		cfg.ResponseTimeBadgePeriods = append([]string(nil), defaultResponseTimeBadgePeriods...)
	}
	seenResponseTimeBadgePeriods := make(map[string]struct{}, len(cfg.ResponseTimeBadgePeriods))
	for _, period := range cfg.ResponseTimeBadgePeriods {
		if period != "30d" && period != "7d" && period != "24h" && period != "1h" {
			return ErrInvalidResponseTimeBadgePeriod
		}
		if _, exists := seenResponseTimeBadgePeriods[period]; exists {
			return ErrInvalidResponseTimeBadgePeriod
		}
		seenResponseTimeBadgePeriods[period] = struct{}{}
	}
	if len(cfg.Favicon.Default) == 0 {
		cfg.Favicon.Default = defaultFavicon
	}
	if len(cfg.Favicon.Size16x16) == 0 {
		cfg.Favicon.Size16x16 = defaultFavicon16
	}
	if len(cfg.Favicon.Size32x32) == 0 {
		cfg.Favicon.Size32x32 = defaultFavicon32
	}
	for _, btn := range cfg.Buttons {
		if err := btn.Validate(); err != nil {
			return err
		}
	}
	// Validate that the template works
	t, err := template.ParseFS(static.FileSystem, static.IndexPath)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	return t.Execute(&buffer, ViewData{UI: cfg, Theme: "dark"})
}

type ViewData struct {
	UI    *Config
	Theme string
}
