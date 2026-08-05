package api

import (
	_ "embed"
	"html/template"
	"net/url"
	"strings"

	"github.com/TwiN/gatus/v5/config/ui"
	static "github.com/TwiN/gatus/v5/web"
	"github.com/TwiN/logr"
	"github.com/gofiber/fiber/v2"
)

func SingleEndpointRoot(uiConfig *ui.Config) fiber.Handler {
	spaHandler := SinglePageApplication(uiConfig)
	return func(c *fiber.Ctx) error {
		if len(uiConfig.SingleEndpoint) > 0 {
			return c.Redirect("/endpoints/"+url.PathEscape(uiConfig.SingleEndpoint), fiber.StatusFound)
		}
		return spaHandler(c)
	}
}

func RestrictToSingleEndpoint(uiConfig *ui.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(uiConfig.SingleEndpoint) == 0 {
			return c.Next()
		}
		key, err := url.QueryUnescape(c.Params("key"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid key encoding")
		}
		if !strings.EqualFold(key, uiConfig.SingleEndpoint) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.Next()
	}
}

func SinglePageApplication(uiConfig *ui.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vd := ui.ViewData{UI: uiConfig}
		{
			themeFromCookie := string(c.Request().Header.Cookie("theme"))
			if len(themeFromCookie) > 0 {
				if themeFromCookie == "dark" {
					vd.Theme = "dark"
				}
			} else if uiConfig.IsDarkMode() { // Since there's no theme cookie, we'll rely on ui.DarkMode
				vd.Theme = "dark"
			}
		}
		t, err := template.ParseFS(static.FileSystem, static.IndexPath)
		if err != nil {
			// This should never happen, because ui.ValidateAndSetDefaults validates that the template works.
			logr.Errorf("[api.SinglePageApplication] Failed to parse template. This should never happen, because the template is validated on start. Error: %s", err.Error())
			return c.Status(500).SendString("Failed to parse template. This should never happen, because the template is validated on start.")
		}
		c.Set("Content-Type", "text/html")
		err = t.Execute(c, vd)
		if err != nil {
			// This should never happen, because ui.ValidateAndSetDefaults validates that the template works.
			logr.Errorf("[api.SinglePageApplication] Failed to execute template. This should never happen, because the template is validated on start. Error: %s", err.Error())
			return c.Status(500).SendString("Failed to parse template. This should never happen, because the template is validated on start.")
		}
		return c.SendStatus(200)
	}
}
