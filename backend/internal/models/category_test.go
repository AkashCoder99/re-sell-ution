package models

import "testing"

func TestBuildCategoryTree(t *testing.T) {
	furnitureID := "root-furniture"
	electronicsID := "root-electronics"

	tree := buildCategoryTree([]Category{
		{ID: "child-laptops", Name: "Laptops", Slug: "laptops", ParentID: &electronicsID},
		{ID: furnitureID, Name: "Furniture", Slug: "furniture"},
		{ID: "child-cameras", Name: "Cameras", Slug: "cameras", ParentID: &electronicsID},
		{ID: electronicsID, Name: "Electronics", Slug: "electronics"},
	})

	if len(tree) != 2 {
		t.Fatalf("expected 2 root categories, got %d", len(tree))
	}
	if tree[0].Name != "Electronics" || tree[1].Name != "Furniture" {
		t.Fatalf("expected sorted root nodes, got %+v", []string{tree[0].Name, tree[1].Name})
	}

	if len(tree[0].Children) != 2 {
		t.Fatalf("expected 2 child categories, got %d", len(tree[0].Children))
	}
	if tree[0].Children[0].Name != "Cameras" || tree[0].Children[1].Name != "Laptops" {
		t.Fatalf("expected sorted child nodes, got %+v", []string{tree[0].Children[0].Name, tree[0].Children[1].Name})
	}
}

func TestIsCanonicalCategorySlug(t *testing.T) {
	if !isCanonicalCategorySlug("mobiles-tablets") {
		t.Fatal("expected canonical OLX slug to be allowed")
	}
	if isCanonicalCategorySlug("electronics") {
		t.Fatal("expected legacy slug to be filtered out")
	}
}
