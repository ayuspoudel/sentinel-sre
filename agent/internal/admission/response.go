package admission

type AdmissionReviewResponse struct {
	Response *AdmissionResponse `json:"response"`
}

type AdmissionResponse struct {
	UID     string  `json:"uid"`
	Allowed bool    `json:"allowed"`
	Result  *Status `json:"status,omitempty"`
}

type Status struct {
	Message string `json:"message"`
}

func (r AdmissionReview) Response(allowed bool, reason string) AdmissionReviewResponse {
	return AdmissionReviewResponse{
		Response: &AdmissionResponse{
			UID:     r.Request.UID,
			Allowed: allowed,
			Result: &Status{
				Message: reason,
			},
		},
	}
}
