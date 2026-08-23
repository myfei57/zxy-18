package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAggregateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string `json:"namespace"`
		TraceID   string `json:"trace_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.hub.Aggregate(req.Namespace, req.TraceID)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAggregateAll(w http.ResponseWriter, r *http.Request) {
	count, err := s.hub.AggregateAll()
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"aggregated": count})
}

func (s *Server) handleAggregateResults(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	writeJSON(w, http.StatusOK, s.hub.AggregateResults(limit))
}

func (s *Server) handleAggregateStatus(w http.ResponseWriter, r *http.Request) {
	cursors := map[string]uint64{}
	for _, namespace := range s.hub.Namespaces() {
		cursors[namespace.Name] = s.hub.AggregateCursor(namespace.Name)
	}
	writeJSON(w, http.StatusOK, map[string]any{"cursors": cursors})
}

func (s *Server) handleOpenWindows(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.OpenWindows())
}

func (s *Server) handleCloseWindow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string `json:"namespace"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	windowID := chi.URLParam(r, "windowID")
	summary, err := s.hub.CloseWindow(req.Namespace, windowID)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleSummaries(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	writeJSON(w, http.StatusOK, s.hub.Summaries(limit))
}
