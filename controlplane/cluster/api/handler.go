package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/events"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/service"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

/*
Store is a type of struct we already have in store.go
We have defined all functions in postgres_store.go because we will be using postgres
It has functions like Create(cluster), Get(clusterName), List(), Delete(clusterName)
Cluster is also a struct defined in model.go
*/
type Handler struct {
	svc        *service.Service
	publisher  events.ClusterIntentPublisher
	kubeclient *kube.KubeClient
}

func NewHandler(svc *service.Service, publisher events.ClusterIntentPublisher, kubeclient *kube.KubeClient) *Handler {
	return &Handler{svc: svc, publisher: publisher, kubeclient: kubeclient}
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

	c := &model.Cluster{Name: req.Name, CredentialRef: req.CredentialRef, Labels: req.Labels}
	err = h.svc.Register(r.Context(), c)

	if err != nil {
		http.Error(w, "failed to register cluster", http.StatusInternalServerError)
		return
	}

	if h.publisher != nil {
		err = h.publisher.PublishClusterRegistered(r.Context(), c)
		if err != nil {
			log.Printf("[REDIS_INFO] failed to publish cluster registered event")
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clusters, err := h.svc.List(r.Context())
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

	name := strings.TrimPrefix(r.URL.Path, "/v1/clusters/")

	if name == "" {
		http.Error(w, "cluster name required", http.StatusBadRequest)
		return
	}

	cluster, err := h.svc.Get(r.Context(), name)
	if err != nil {
		http.Error(w, "failed to get cluster", http.StatusInternalServerError)
		return
	}
	if cluster == nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cluster)
}

func (h *Handler) DeleteByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/v1/clusters/")
	if name == "" {
		http.Error(w, "cluster name required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	cluster, err := h.svc.Get(ctx, name)
	if err != nil {
		http.Error(w, "failed to get cluster", http.StatusInternalServerError)
		return
	}
	if cluster == nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	if err := h.svc.Delete(ctx, name); err != nil {
		http.Error(w, "failed to delete cluster", http.StatusInternalServerError)
		return
	}

	if h.kubeclient != nil {
		const namespace = "sentinel"
		secretName := cluster.CredentialRef

		if err := h.kubeclient.DeleteKubeconfigSecret(ctx, namespace, secretName); err != nil {
			log.Printf("[WARN] failed to delete kubeconfig secret %s/%s: %v",
				namespace, secretName, err)
		}
	}
	if h.publisher != nil {
		if err := h.publisher.PublishClusterDeleted(ctx, name); err != nil {
			log.Printf("[REDIS_INFO] failed to publish cluster deleted event")
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// DB reachability
	_, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

/*
@ayuspoudel
RegisgterWithCredentials handler allows users to register their cluster to sentinel
by providing path to a kubeconfig file in a API call.
It does two things it validates the kubeconfig:
  - If multiple contexts are present in the cluster
  - See if contextName was provided during API call
  - If provided we check if that context exists in kubeconfig
  - If not provided we check
  - If clusterName from API call matches the contextName existing in kubeconfig
  - If matches we use that context
  - If not matches we throw error to user to provide contextName

Validation solves the edge case when users apply kubeconfigs with multiple cluster contexts
After validation it takes the kubeconfig and applies it as a secret into sentinel namespace
with secret name sentinel-cluster-{clusterName} (clusterName is saturated into k8s friendly
by our function in ./naming.go)
*/
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

	//	VALIDATION PHASE

	cfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		http.Error(w, "invalid kubeconfig: "+err.Error(), http.StatusBadRequest)
		return
	}
	contextName := r.FormValue("context")
	if contextName == "" {
		// If contextName is not provided we check if clusterName matches any context in kubeconfig
		if len(cfg.Contexts) == 1 {
			for k := range cfg.Contexts {
				contextName = k
				break
			}
		} else {
			if _, ok := cfg.Contexts[name]; ok {
				contextName = name
			}
		}
	}
	_, ok := cfg.Contexts[contextName]
	if !ok {
		http.Error(w, fmt.Sprintf("context %q not found in kubeconfig (avaliable: %v)", contextName, contextKeys(cfg.Contexts)), http.StatusBadRequest)
		return
	}
	// We have validated the kubeconfig and contextName provided by user
	safeName := model.ToDNS1123Name(name)
	secretName := fmt.Sprintf("sentinel-cluster-%s", safeName)
	namespace := "sentinel"
	err = h.kubeclient.EnsureKubeconfigSecret(r.Context(), namespace, secretName, kubeconfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cluster := &model.Cluster{
		Name:          name,
		CredentialRef: secretName,
		Labels:        map[string]string{"context": contextName},
	}

	if err := h.svc.Register(r.Context(), cluster); err != nil {
		http.Error(w, "failed to register cluster", http.StatusInternalServerError)
		return
	}
	if h.publisher != nil {
		err = h.publisher.PublishClusterRegistered(r.Context(), cluster)
		if err != nil {
			log.Printf("[REDIS_INFO] failed to publish cluster registered event")
		}
	}

	resp := ClusterResponse{
		ClusterName:   cluster.Name,
		CredentialRef: cluster.CredentialRef,
		Labels:        cluster.Labels,
		RegisteredAt:  time.Now().UTC().Format(time.RFC3339),
		Source:        "api",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)

}

/*
@ayuspoudel
This is a small helper function that converts a map of {contextName : *context}
into a array of strings that only contain contextNames
*/
func contextKeys(m map[string]*clientcmdapi.Context) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
