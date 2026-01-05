package cluster

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

/*
@ayuspoudel
This is is a register request struct. A struct type that defines how a request to
register a struct should come in as. This allows us to directly do json.Encode(&request)
and help validate the request easily.
*/
type RegisterRequest struct {
	Name    string `json:"name"`
	AgentID string `json:"agent_id"`
	Version string `json:"version"`
	Address string `json:"address"`
}

func RegisterHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.AgentID == "" {
			http.Error(w, "name and agent_id required", http.StatusBadRequest)
			return
		}

		store.Register(&Cluster{
			Name:    req.Name,
			AgentID: req.AgentID,
			Version: req.Version,
			Address: req.Address,
		})

		log.Printf("[cluster-register] Name=%s AgentId=%s Version=%s Address=%s", req.Name, req.AgentID, req.Version, req.Address)

		w.WriteHeader(http.StatusCreated)
	}
}

func ListHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		_ = json.NewEncoder(w).Encode(store.List())
	}
}

func GetByNameHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		/*
			@ayuspoudel
			We extract cluster name from the URL path.
			Expected format: /clusters/{name}
		*/
		name := strings.TrimPrefix(r.URL.Path, "/clusters/")
		if name == "" {
			http.Error(w, "cluster name required", http.StatusBadRequest)
			return
		}

		cluster, ok := store.Get(name)
		if !ok {
			http.Error(w, "cluster not found", http.StatusNotFound)
			return
		}

		_ = json.NewEncoder(w).Encode(cluster)
	}
}
