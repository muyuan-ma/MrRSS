package digest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleListDigests returns recent daily AI digests.
func HandleListDigests(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	limit := 20
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	digests, err := h.DB.GetDailyDigests(limit)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, digests)
}

// HandleLatestDigest returns the latest daily AI digest with article summaries.
func HandleLatestDigest(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	digest, err := h.DB.GetLatestDailyDigest()
	if err != nil {
		if err == sql.ErrNoRows {
			response.JSON(w, map[string]interface{}{"digest": nil})
			return
		}
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	digest.Articles, _ = h.DB.GetDailyDigestArticles(digest.ID)
	response.JSON(w, digest)
}

// HandleGenerateDigest triggers manual daily AI digest generation.
func HandleGenerateDigest(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Force bool `json:"force"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	digest, err := h.GenerateDailyDigest(ctx, req.Force)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, digest)
}

// HandleMarkDigestNotified records that the latest digest has been shown to the user.
func HandleMarkDigestNotified(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}
	if req.ID <= 0 {
		response.Error(w, errors.New("invalid digest id"), http.StatusBadRequest)
		return
	}

	if err := h.DB.MarkDailyDigestNotified(req.ID); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]bool{"success": true})
}
