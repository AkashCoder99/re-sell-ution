package handlers

import (
	"context"
	"net/http"

	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

type categoryStore interface {
	List(ctx context.Context) ([]models.Category, error)
	Tree(ctx context.Context) ([]models.CategoryTreeNode, error)
}

type CategoryHandler struct {
	Categories categoryStore
}

func (h CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Categories.List(r.Context())
	if err != nil {
		observability.Error(r.Context(), "categories.list.failed", map[string]any{"error": err.Error()})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load categories"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"categories": categories})
}

func (h CategoryHandler) Tree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.Categories.Tree(r.Context())
	if err != nil {
		observability.Error(r.Context(), "categories.tree.failed", map[string]any{"error": err.Error()})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load category tree"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"categories": tree})
}
