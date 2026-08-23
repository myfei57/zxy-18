package console

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	errMissingNamespace = errors.New("namespace query parameter is required")
	errSpanNotFound     = errors.New("span not found")
)

func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.Namespaces())
}

func (s *Server) handleRegisterNamespace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Service string `json:"service"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	namespace, err := s.hub.RegisterNamespace(req.Name, req.Service)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, namespace)
}

func (s *Server) handleDeactivateNamespace(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	if err := s.hub.DeactivateNamespace(namespace); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

func (s *Server) handleIngestSpan(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.StartedAt == 0 {
		req.StartedAt = time.Now().Unix()
	}
	sp, decided, err := s.hub.IngestSpan(req)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"span":    toSpanView(sp, 0),
		"sampled": decided,
	})
}

func (s *Server) handleSpanList(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		writeError(w, http.StatusBadRequest, errMissingNamespace)
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	writeJSON(w, http.StatusOK, s.hub.SpanList(namespace, limit))
}

func (s *Server) handleSpanDetail(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	spanID := chi.URLParam(r, "spanID")
	detail, ok := s.hub.SpanDetail(namespace, spanID)
	if !ok {
		writeError(w, http.StatusNotFound, errSpanNotFound)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleFinishSpan(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	spanID := chi.URLParam(r, "spanID")
	var req struct {
		Payload map[string]string `json:"payload"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.hub.FinishSpan(namespace, spanID, req.Payload); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "finished"})
}
