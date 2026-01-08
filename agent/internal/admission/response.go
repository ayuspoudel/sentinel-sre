package admission

type AdmissionReviewResponse struct {
	Response *AdmissionResponse `json:"response"`
}

// Sending response is simple we will simply send Allowed and Result with UID inside reponse:{} json
type AdmissionResponse struct {
	UID     string  `json:"uid"`
	Allowed bool    `json:"allowed"`
	Result  *Status `json:"status,omitempty"`
}

// WE made result a nested json which will have a simple message inside of it
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
