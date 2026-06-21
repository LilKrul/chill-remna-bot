package web

import (
	"io/fs"
	"net/http"
)

// handleMiniStatic serves the embedded Mini App SPA under /miniapp/.
// Gated on the feature flag.
func (s *Server) handleMiniStatic(w http.ResponseWriter, r *http.Request) {
	if s.mini == nil || !s.mini.MiniEnabled() {
		http.NotFound(w, r)
		return
	}
	sub, err := fs.Sub(miniStaticFS, "miniapp_static")
	if err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	http.StripPrefix("/miniapp/", http.FileServer(http.FS(sub))).ServeHTTP(w, r)
}
