package admission

import (
	"encoding/json"
	"net/http"

	"github.com/ayuspoudel/sentinel-sre/agent/client"
)

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var review AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if review.Request == nil {
		http.Error(w, "invalid admission review", http.StatusBadRequest)
		return
	}

	guardName, err := extractGuardName(review.Request)
	if err != nil {
		resp := review.Response(false, "failed to extract guard: "+err.Error())
		writeResponse(w, resp)
		return
	}

	if guardName == "" {
		resp := review.Response(false, "missing sentinel.guard label")
		writeResponse(w, resp)
		return
	}

	allowed, reason := client.CheckWithSentinel(r.Context(), guardName)

	resp := review.Response(allowed, reason)
	writeResponse(w, resp)
}

func writeResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
