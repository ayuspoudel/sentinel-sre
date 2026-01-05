package webhook

import (
	"encoding/json"
	"net/http"
)

type AdmissionResponse struct {
	UID     string  `json:"uid"`
	Allowed bool    `json:"allowed"`
	Status  *Status `json:"status,omitempty"`
}

type Status struct {
	Message string `json:"message"`
}

type AdmissionReviewResponse struct {
	Response *AdmissionResponse `json:"response"`
}

func WriteAllow(w http.ResponseWriter, uid string) {
	resp := AdmissionReviewResponse{
		Response: &AdmissionResponse{
			UID:     uid,
			Allowed: true,
		},
	}
	write(w, resp)
}

func WriteDeny(w http.ResponseWriter, uid, reason string) {
	resp := AdmissionReviewResponse{
		Response: &AdmissionResponse{
			UID:     uid,
			Allowed: false,
			Status: &Status{
				Message: reason,
			},
		},
	}

	write(w, resp)
}

func write(w http.ResponseWriter, resp AdmissionReviewResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
