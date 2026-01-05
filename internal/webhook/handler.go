package webhook

import (
	"net/http"

	"github.com/ayuspoudel/sentinel-sre/internal/action"
	"github.com/ayuspoudel/sentinel-sre/internal/engine"
)

type Handler struct {
	engine *engine.Engine
}

func NewHandler(engine *engine.Engine) *Handler {
	return &Handler{
		engine: engine,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	review, err := DecodeAdmissionReview(r)
	if err != nil {
		WriteDeny(w, "", "invalid admission review")
		return
	}

	guard := ExtractGuardName(review)
	if guard == "" {
		WriteDeny(w, review.Request.UID, "missing sentinel.guard/name annotation")
		return
	}

	act, ok := h.engine.Actions().Get(guard)
	if !ok {
		WriteDeny(w, review.Request.UID, "no sentinel decision available")
		return
	}

	if act.Type == action.Block {
		WriteDeny(w, review.Request.UID, act.Reason)
		return
	}

	WriteAllow(w, review.Request.UID)
	return
}
