package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

type Service struct {
	Handler *Handler
	Mux     *http.ServeMux
}

func NewService(gatewayURL, productURL, orderURL string) *Service {
	handler := NewHandler(gatewayURL, productURL, orderURL)
	mux := http.NewServeMux()

	staticSub, _ := fs.Sub(staticFiles, "static")
	staticFS := http.FileServer(http.FS(staticSub))

	RegisterRoutes(mux, handler, staticFS)

	return &Service{
		Handler: handler,
		Mux:     mux,
	}
}
