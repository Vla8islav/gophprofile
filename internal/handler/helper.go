package handler

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

// decodeCredentials: parses the request body and returns the login and password.
func (h *Handler) decodeCredentials(w http.ResponseWriter, r *http.Request) (login, password string, ok bool) {
	if r.Method != http.MethodPost {
		h.writeMethodNotAllowed(w, "only POST method is allowed")
		return "", "", false
	}
	mimeType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mimeType != "application/json" {
		h.writeBadRequest(w, "only application/json content type is supported")
		return "", "", false
	}
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeBadRequest(w, "failed to read request body: "+err.Error())
		return "", "", false
	}
	var creds domain.UserLoginRequest // {login, password}
	if err := json.Unmarshal(requestBody, &creds); err != nil {
		h.writeBadRequest(w, "couldn't parse requestBody:"+err.Error())
		return "", "", false
	}
	if creds.Login == "" {
		h.writeBadRequest(w, "login cannot be empty")
		return "", "", false
	}
	if creds.Password == "" {
		h.writeBadRequest(w, "password cannot be empty")
		return "", "", false
	}
	return creds.Login, creds.Password, true
}

// writeToken: the 200 + JSON token response.
func (h *Handler) writeToken(w http.ResponseWriter, token string) {
	payload, err := json.Marshal(domain.UserLoginResponse{Token: token})
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
