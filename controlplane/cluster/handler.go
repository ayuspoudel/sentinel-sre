package cluster

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

/*
Store is a type of struct we already have in store.go
We have defined all functions in postgres_store.go because we will be using postgres
It has functions like Create(cluster), Get(clusterName), List(), Delete(clusterName)
Cluster is also a struct defined in model.go
*/
type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	/*
		RegisterRequest is a struct in api.go it has json: name, credentials_ref, labels
		If a payload fails to provide these three things, it will trigger error
	*/
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.CredentialRef == "" {
		http.Error(w, "name and credential_ref required", http.StatusBadRequest)
		return
	}

	c := &Cluster{Name: req.Name, CredentialRef: req.CredentialRef, Labels: req.Labels}
	err = h.store.Create(r.Context(), c)
	if err != nil {
		http.Error(w, "failed to register cluster", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clusters, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list clusters", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(clusters)
}

func (h *Handler) GetByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/clusters/")
	if name == "" {
		http.Error(w, "cluster name required", http.StatusBadRequest)
		return
	}

	cluster, err := h.store.Get(r.Context(), name)
	if err != nil {
		http.Error(w, "failed to get cluster", http.StatusInternalServerError)
		return
	}
	if cluster == nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(cluster)
}

func (h *Handler) DeleteByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/clusters/")
	if name == "" {
		http.Error(w, "cluster name required", http.StatusBadRequest)
		return
	}

	err := h.store.Delete(r.Context(), name)
	if err != nil {
		http.Error(w, "failed to delete cluster", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// DB reachability
	_, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) RegisterWithCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("kubeconfig")
	if err != nil {
		http.Error(w, "failed to get kubeconfig file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	kubeconfig, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read kubeconfig file", http.StatusInternalServerError)
		return
	}
	safeName := ToDNS1123Name(name)
	secretName := fmt.Sprintf("sentinel-cluster-%s", safeName)
	namespace := "sentinel"
	kubeclient, err := NewKubeClient()
	if err != nil {
		http.Error(w, "failed to init kube client", http.StatusInternalServerError)
		return
	}
	err = EnsureKubeconfigSecret(r.Context(), kubeclient, namespace, secretName, kubeconfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cluster := &Cluster{Name: name, CredentialRef: secretName, Labels: map[string]string{}}
	if err := h.store.Create(r.Context(), cluster); err != nil {
		http.Error(w, "failed to register cluster", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
