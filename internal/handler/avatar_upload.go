package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/audit"
	"github.com/Vla8islav/gophprofile/internal/domain"
)

// multipartOverhead is headroom on top of the file size cap
const multipartOverhead = 1 << 20

// AvatarUploadHandler godoc
// @Summary   Upload an avatar (multipart/form-data, field "file")
// @Tags      avatars
// @Accept    multipart/form-data
// @Produce   json
// @Param     file formData file true "image file (jpeg/png/webp, max 10MB)"
// @Success   201 {object} domain.AvatarUploadResponse
// @Failure   400 {object} handler.apiError
// @Failure   401
// @Failure   413 {object} handler.apiError
// @Failure   500
// @Security  BearerAuth
// @Router    /api/v1/avatars [post]
func (h *Handler) AvatarUploadHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "avatar.upload")

	userID, ok := h.requestUserID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxAvatarSizeBytes+multipartOverhead)
	if err := r.ParseMultipartForm(domain.MaxAvatarSizeBytes + multipartOverhead); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			h.writeJSONError(w, http.StatusRequestEntityTooLarge, apiError{
				Error:   "File too large",
				MaxSize: domain.MaxAvatarSizeBytes,
			})
			return
		}
		h.writeJSONError(w, http.StatusBadRequest, apiError{
			Error:   "Invalid request",
			Details: "expected multipart/form-data with a \"file\" field",
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeJSONError(w, http.StatusBadRequest, apiError{
			Error:   "Invalid request",
			Details: "missing \"file\" field",
		})
		return
	}
	defer func() { _ = file.Close() }()

	if header.Size > domain.MaxAvatarSizeBytes {
		h.writeJSONError(w, http.StatusRequestEntityTooLarge, apiError{
			Error:   "File too large",
			MaxSize: domain.MaxAvatarSizeBytes,
		})
		return
	}

	// Detect the real format from magic bytes; the client-supplied
	// Content-Type header is not trusted.
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		h.writeInternalServerError(w, "failed to read uploaded file: "+err.Error())
		return
	}
	mimeType := http.DetectContentType(head[:n])
	if !domain.AllowedAvatarMimeTypes[mimeType] {
		h.writeJSONError(w, http.StatusBadRequest, apiError{
			Error:   "Invalid file format",
			Details: "Supported formats: jpeg, png, webp",
		})
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		h.writeInternalServerError(w, "failed to rewind uploaded file: "+err.Error())
		return
	}

	avatar, err := h.service.UploadAvatar(r.Context(), userID, header.Filename, mimeType, header.Size, file)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, domain.AvatarUploadResponse{
		ID:        avatar.ID,
		UserID:    avatar.UserID,
		URL:       domain.AvatarURL(avatar.ID),
		Status:    avatar.ProcessingStatus,
		CreatedAt: avatar.CreatedAt,
	})
}
