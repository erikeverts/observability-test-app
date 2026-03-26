package gateway

import (
	"io"
	"net/http"
)

type Handler struct {
	productServiceURL string
	orderServiceURL   string
	httpClient        *http.Client
}

func NewHandler(productServiceURL, orderServiceURL string, httpClient *http.Client) *Handler {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Handler{
		productServiceURL: productServiceURL,
		orderServiceURL:   orderServiceURL,
		httpClient:        httpClient,
	}
}

func (h *Handler) ProxyProducts(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, h.productServiceURL)
}

func (h *Handler) ProxyOrders(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, h.orderServiceURL)
}

func (h *Handler) proxy(w http.ResponseWriter, r *http.Request, targetBase string) {
	targetURL := targetBase + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "failed to create proxy request", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
