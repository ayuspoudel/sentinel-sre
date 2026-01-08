package admission

import (
	"encoding/json"
	"net/http"

	"github.com/ayuspoudel/sentinel-sre/agent/internal/client"
)

type Handler struct {
	sentinel *client.SentinelClient
}

func NewHandler(s *client.SentinelClient) *Handler {
	return &Handler{sentinel: s}
}
func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	// if the method is not a post request we deny it because its none of our webhook server's interest to serve a get... request
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed ", http.StatusMethodNotAllowed)
		return
	}
	// we build a admission review struct so we can decode the json we recieved as payload into it from kubernetes api
	var review AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// if k8s api sends something ineligible to become our request we deny it
	if review.Request == nil {
		http.Error(w, "invalid admission review", http.StatusBadRequest)
		return
	}
	// Now we are extracting guard name from the request, because our webhook server will send the request to main sentinel control plane which understands in term of guard name
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
	// We are here calling sentinel api to check if the guard for this application allows it to be deployed right now or not
	allowed, reason := h.sentinel.CheckWithSentinel(r.Context(), guardName)

	// We get the resposne from sentinel, and feed it into our already built response struct inside Admission review and we write the response
	// reponse function already returns a AdmissionReviewResponse struct and we feed allowed and reason in it
	// it is a method of AdmissionReview and will have access to UID already
	resp := review.Response(allowed, reason)
	writeResponse(w, resp)
}

func writeResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

/*
GuardName will be as a label inside metadata
So in a request it will come as:

	request:{
		kind:...
		namespace:...
		name:...
		object:{
		   metadata: {
		   		name: ...
				labels:{
						sentinel.guard: GUARDNAME  <------------- THIS IS WHAT WE NEED
					}
				}
			}
		}
	}
*/
func extractGuardName(req *AdmissionRequest) (string, error) {
	if req == nil {
		return "", nil
	}

	var obj RawObject
	if err := json.Unmarshal(req.Object, &obj); err != nil {
		return "", err
	}

	return obj.Metadata.Labels["sentinel.guard"], nil
}
