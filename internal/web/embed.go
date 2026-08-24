// Package web embeds the four TraceFlow control pages.
package web

import _ "embed"

//go:embed spans.html
var SpansHTML string

//go:embed report.html
var ReportHTML string

//go:embed aggregate.html
var AggregateHTML string

//go:embed audit.html
var AuditHTML string
