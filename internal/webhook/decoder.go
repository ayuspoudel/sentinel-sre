package webhook

import (
	"encoding/json"
	"net/http"
)

type AdmissionReview struct {
	Request *AdmissionRequest `json:"request"`
}

type AdmissionRequest struct {
	UID       string          `json:"uid"`
	Operation string          `json:"operation"`
	Object    json.RawMessage `json:"object"`
}

func DecodeAdmissionReview(r *http.Request) (*AdmissionReview, error) {
	defer r.Body.Close()
	var review AdmissionReview
	err := json.NewDecoder(r.Body).Decode(&review)
	if err != nil {
		return nil, err
	}
	if review.Request == nil {
		return nil, nil
	}
	return &review, nil
}
