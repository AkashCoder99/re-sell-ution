package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"resellution/backend/internal/models"
)

type stubCategoryStore struct {
	listFn func(ctx context.Context) ([]models.Category, error)
	treeFn func(ctx context.Context) ([]models.CategoryTreeNode, error)
}

func (s stubCategoryStore) List(ctx context.Context) ([]models.Category, error) {
	if s.listFn == nil {
		return []models.Category{}, nil
	}
	return s.listFn(ctx)
}

func (s stubCategoryStore) Tree(ctx context.Context) ([]models.CategoryTreeNode, error) {
	if s.treeFn == nil {
		return []models.CategoryTreeNode{}, nil
	}
	return s.treeFn(ctx)
}

func TestCategoryHandlerListOK(t *testing.T) {
	h := CategoryHandler{
		Categories: stubCategoryStore{
			listFn: func(ctx context.Context) ([]models.Category, error) {
				return []models.Category{{ID: "c1", Name: "Electronics", Slug: "electronics-appliances"}}, nil
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCategoryHandlerListFailure(t *testing.T) {
	h := CategoryHandler{
		Categories: stubCategoryStore{
			listFn: func(ctx context.Context) ([]models.Category, error) {
				return nil, errors.New("db down")
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestCategoryHandlerTreeOK(t *testing.T) {
	h := CategoryHandler{
		Categories: stubCategoryStore{
			treeFn: func(ctx context.Context) ([]models.CategoryTreeNode, error) {
				return []models.CategoryTreeNode{{ID: "root", Name: "Electronics"}}, nil
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/tree", nil)
	h.Tree(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCategoryHandlerTreeFailure(t *testing.T) {
	h := CategoryHandler{
		Categories: stubCategoryStore{
			treeFn: func(ctx context.Context) ([]models.CategoryTreeNode, error) {
				return nil, errors.New("db down")
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/tree", nil)
	h.Tree(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

