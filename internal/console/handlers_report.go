package console

import "net/http"

func (s *Server) handleFlush(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string `json:"namespace"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	batch, err := s.hub.Flush(req.Namespace)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string `json:"namespace"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	batch, err := s.hub.Send(req.Namespace)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (s *Server) handleReportStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.ReportStatus())
}
