package dashboard

import (
	"encoding/json"
	"io"
	"net/http"
)

type ServiceInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Handler struct {
	services   []ServiceInfo
	httpClient *http.Client
}

func NewHandler(gatewayURL, productURL, orderURL, inventoryURL string) *Handler {
	return &Handler{
		services: []ServiceInfo{
			{Name: "gateway", URL: gatewayURL},
			{Name: "product", URL: productURL},
			{Name: "order", URL: orderURL},
			{Name: "inventory", URL: inventoryURL},
		},
		httpClient: &http.Client{},
	}
}

func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.services)
}

func (h *Handler) GetServiceChaos(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	svc := h.findService(name)
	if svc == nil {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}

	resp, err := h.httpClient.Get(svc.URL + "/admin/chaos")
	if err != nil {
		http.Error(w, `{"error":"service unreachable: `+err.Error()+`"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handler) SetServiceChaos(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	svc := h.findService(name)
	if svc == nil {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPut, svc.URL+"/admin/chaos", r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"service unreachable: `+err.Error()+`"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handler) ClearServiceDisk(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	svc := h.findService(name)
	if svc == nil {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, svc.URL+"/admin/chaos/clear-disk", nil)
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"service unreachable: `+err.Error()+`"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handler) findService(name string) *ServiceInfo {
	for i := range h.services {
		if h.services[i].Name == name {
			return &h.services[i]
		}
	}
	return nil
}
