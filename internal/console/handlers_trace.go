package console

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"traceflow/internal/trace"
)

var errTraceNotFound = errors.New("trace not found")

func (s *Server) handleRecentTraces(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state != "" && !trace.Valid(state) {
		writeError(w, http.StatusBadRequest, errors.New("unknown trace state filter"))
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	writeJSON(w, http.StatusOK, s.hub.RecentTraces(limit))
}

func (s *Server) handleTraceQuery(w http.ResponseWriter, r *http.Request) {
	traceID := chi.URLParam(r, "traceID")
	view, ok := s.hub.QueryTrace(traceID)
	if !ok {
		writeError(w, http.StatusNotFound, errTraceNotFound)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleTraceTree(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	traceID := chi.URLParam(r, "traceID")
	nodes, ok := s.hub.TraceTree(namespace, traceID)
	if !ok {
		writeError(w, http.StatusNotFound, errTraceNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trace_id": traceID,
		"depth":    trace.Depth(nodes),
		"nodes":    nodes,
	})
}
