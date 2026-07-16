package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"postsapi/internal/models"
	"postsapi/internal/validation"
)

type PostHandler struct {
	Repo *models.PostRepository
}

func NewPostHandler(repo *models.PostRepository) *PostHandler {
	return &PostHandler{Repo: repo}
}

// ---- response helpers -------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"message": message,
	})
}

func writeValidationErrors(w http.ResponseWriter, errs validation.Errors) {
	writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
		"success": false,
		"message": "validation failed",
		"errors":  errs,
	})
}

// ---- endpoints ----------------------------------------------------------

// Create handles: POST /article/
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.PostInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if errs := validation.ValidatePostInput(in); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))

	post, err := h.Repo.Create(in)
	if err != nil {

		writeError(w, http.StatusInternalServerError, "failed to create article")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    post,
	})
}

// List handles: GET /article/{limit}/{offset}
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.PathValue("limit"))
	if err != nil || limit <= 0 {
		writeError(w, http.StatusBadRequest, "limit must be a positive integer")
		return
	}
	offset, err := strconv.Atoi(r.PathValue("offset"))
	if err != nil || offset < 0 {
		writeError(w, http.StatusBadRequest, "offset must be a non-negative integer")
		return
	}

	// Guard against unbounded queries.
	if limit > 200 {
		limit = 200
	}

	posts, total, err := h.Repo.List(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch articles")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    posts,
		"meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// Get handles: GET /article/{id}
func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return
	}

	post, err := h.Repo.GetByID(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch article")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    post,
	})
}

// Update handles: PUT/PATCH /article/{id}
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return
	}

	var in models.PostInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if errs := validation.ValidatePostInput(in); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))

	post, err := h.Repo.Update(id, in)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update article")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    post,
	})
}

// Delete handles: DELETE /article/{id}
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return
	}

	err = h.Repo.Delete(id)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete article")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "article deleted",
	})
}
