package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "traceflow"})
}

func (s *Server) handleIndexBuild(w http.ResponseWriter, r *http.Request) {
	number, err := s.hub.BuildIndex()
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"generation": number})
}

func (s *Server) handleIndexStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.IndexStatus())
}

func (s *Server) handleIndexSnapshot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.IndexSnapshot())
}

func (s *Server) handleSampleRate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Rate int `json:"rate"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.hub.SwitchRate(req.Rate); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, s.hub.SamplerStatus())
}

func (s *Server) handleSampleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.SamplerStatus())
}

func (s *Server) handleQuotaList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.QuotaList())
}

func (s *Server) handleSetQuota(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	var req struct {
		Limit int `json:"limit"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.hub.SetQuota(namespace, req.Limit); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, s.hub.QuotaStatus(namespace))
}

func (s *Server) handleAuditRecent(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	writeJSON(w, http.StatusOK, s.hub.AuditRecent(limit))
}

func (s *Server) handleAuditStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.AuditStats())
}

func (s *Server) handleMaintenance(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.Maintenance(); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "maintenance complete"})
}
