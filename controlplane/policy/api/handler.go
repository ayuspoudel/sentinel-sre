package api

import (
	"encoding/json"
	"net/http"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/service"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
)

type Handler struct {
	svc service.PolicyService
}

func NewHandler(svc service.PolicyService) *Handler {
	return &Handler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, ErrorResponse{Error: err.Error()})
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) ApplyPolicy(w http.ResponseWriter, r *http.Request, name string) {
	var req ApplyPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	p := &spec.PolicySpec{
		Metadata: req.Metadata,
		Target:   req.Target,
		Signals:  req.Signals,
		Policy:   req.Policy,
	}
	p.Metadata.Name = name

	_, err := h.svc.ApplyPolicy(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, ApplyPolicyResponse{Status: "applied"})
}

func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request, name string) {
	p, err := h.svc.GetPolicy(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.svc.ListPolicies(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request, name string) {
	st, err := h.svc.GetStatus(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (h *Handler) DeletePolicy(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.svc.DeletePolicy(r.Context(), name); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
