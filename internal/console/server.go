package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"traceflow/internal/web"
)

// Server serves the console pages and JSON API.
type Server struct {
	hub *Hub
}

// NewServer creates a console server.
func NewServer(hub *Hub) *Server {
	return &Server{hub: hub}
}

// Handler builds the chi router.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.handleIndex)
	r.Get("/spans", s.handlePage(web.SpansHTML))
	r.Get("/report", s.handlePage(web.ReportHTML))
	r.Get("/aggregate", s.handlePage(web.AggregateHTML))
	r.Get("/audit", s.handlePage(web.AuditHTML))
	r.Get("/api/health", s.handleHealth)
	r.Get("/api/namespaces", s.handleNamespaces)
	r.Post("/api/namespaces", s.handleRegisterNamespace)
	r.Delete("/api/namespaces/{namespace}", s.handleDeactivateNamespace)
	r.Get("/api/spans", s.handleSpanList)
	r.Post("/api/spans", s.handleIngestSpan)
	r.Get("/api/spans/{namespace}/{spanID}", s.handleSpanDetail)
	r.Post("/api/spans/{namespace}/{spanID}/finish", s.handleFinishSpan)
	r.Get("/api/traces", s.handleRecentTraces)
	r.Get("/api/traces/{traceID}", s.handleTraceQuery)
	r.Get("/api/traces/{namespace}/{traceID}/tree", s.handleTraceTree)
	r.Post("/api/report/flush", s.handleFlush)
	r.Post("/api/report/send", s.handleSend)
	r.Get("/api/report/status", s.handleReportStatus)
	r.Post("/api/aggregate/run", s.handleAggregateRun)
	r.Post("/api/aggregate/all", s.handleAggregateAll)
	r.Get("/api/aggregate/results", s.handleAggregateResults)
	r.Get("/api/aggregate/status", s.handleAggregateStatus)
	r.Get("/api/aggregate/windows", s.handleOpenWindows)
	r.Post("/api/aggregate/windows/{windowID}/close", s.handleCloseWindow)
	r.Get("/api/aggregate/summaries", s.handleSummaries)
	r.Post("/api/index/build", s.handleIndexBuild)
	r.Get("/api/index/status", s.handleIndexStatus)
	r.Get("/api/index/snapshot", s.handleIndexSnapshot)
	r.Post("/api/sample/rate", s.handleSampleRate)
	r.Get("/api/sample/status", s.handleSampleStatus)
	r.Get("/api/quota", s.handleQuotaList)
	r.Put("/api/quota/{namespace}", s.handleSetQuota)
	r.Get("/api/audit", s.handleAuditRecent)
	r.Get("/api/audit/stats", s.handleAuditStats)
	r.Post("/api/maintenance", s.handleMaintenance)
	return r
}

// Start serves the console until the process exits.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
