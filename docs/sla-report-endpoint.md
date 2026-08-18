# SLA Report Endpoint

Gatus generates an on-demand PDF SLA report at:

```text
GET /api/v1/reports/sla.pdf
```

The report is intentionally not shown in the web interface. Maintainers request
the URL directly using the dedicated super-admin credentials.

## Configuration

Enable PDF reporting and configure credentials that are different from the
application-wide `security.basic` credentials:

```yaml
reporting:
  pdf:
    enabled: true
    super-admin:
      username: "super-admin"
      password-bcrypt-base64: "<base64-encoded bcrypt hash>"
```

If `super-admin` is omitted, the endpoint fails closed with `403 Forbidden`.
The global Basic Auth username and password cannot download a report.

## Authentication

Use HTTP Basic Auth with the `reporting.pdf.super-admin` account. The endpoint
challenges clients with the `SLA Reports` realm.

```bash
curl -u 'super-admin:password' \
  -o service-sla-report.pdf \
  'http://localhost:8080/api/v1/reports/sla.pdf?period=30d'
```

`curl` prompts for the password. A browser requests the same credentials when
opening the URL directly.

## Reporting Period

Choose one reporting-period form:

| Query parameters | Example | Notes |
| --- | --- | --- |
| `period` | `?period=30d` | Allowed values must be enabled in `reporting.pdf.allowed-periods`. Supported values are `90d`, `60d`, `30d`, `7d`, `1d`, and `1h`. |
| `from` and `to` | `?from=2026-07-19&to=2026-08-18` | Date range of up to 90 days. `start` and `end` are accepted aliases. |
| Date range in `period` | `?period=2026-07-19..2026-08-18` | The separators `..`, `_`, `:`, and `/` are accepted. |

For `from` and `to`, dates may use `YYYY-MM-DD`, a local date-time, or
RFC3339. Date-only values use the configured `reporting.pdf.timezone`; the end
date includes the whole day.

## Responses

| Status | Meaning |
| --- | --- |
| `200 OK` | PDF report returned as an attachment. |
| `400 Bad Request` | Invalid, incomplete, disabled, or more-than-90-day period. |
| `401 Unauthorized` | Missing or invalid super-admin credentials. |
| `403 Forbidden` | Super-admin report credentials are not configured. |
| `404 Not Found` | PDF reporting is disabled. |
