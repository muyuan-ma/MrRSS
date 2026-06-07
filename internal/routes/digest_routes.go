package routes

import (
	"net/http"

	"MrRSS/internal/handlers/core"
	digesthandlers "MrRSS/internal/handlers/digest"
)

// registerDigestRoutes registers AI agent daily digest routes.
func registerDigestRoutes(mux *http.ServeMux, h *core.Handler) {
	mux.HandleFunc("/api/agent/digests", func(w http.ResponseWriter, r *http.Request) {
		digesthandlers.HandleListDigests(h, w, r)
	})
	mux.HandleFunc("/api/agent/digest/latest", func(w http.ResponseWriter, r *http.Request) {
		digesthandlers.HandleLatestDigest(h, w, r)
	})
	mux.HandleFunc("/api/agent/digest/generate", func(w http.ResponseWriter, r *http.Request) {
		digesthandlers.HandleGenerateDigest(h, w, r)
	})
	mux.HandleFunc("/api/agent/digest/notified", func(w http.ResponseWriter, r *http.Request) {
		digesthandlers.HandleMarkDigestNotified(h, w, r)
	})
}
