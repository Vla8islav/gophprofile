package handler

import (
	"net/http"

	"github.com/Vla8islav/gophprofile/web"
)

// serveWebFile serves one embedded asset with a fixed content type.
func (h *Handler) serveWebFile(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := web.Static.ReadFile("static/" + name)
		if err != nil {
			h.writeInternalServerError(w, "missing embedded asset "+name)
			return
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write(data)
	}
}
