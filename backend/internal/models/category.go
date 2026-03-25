package models

import (
	"context"
	"database/sql"
	"sort"
)

var canonicalCategorySlugs = map[string]struct{}{
	"mobiles-tablets":        {},
	"electronics-appliances": {},
	"furniture":              {},
	"home-garden":            {},
	"fashion":                {},
	"vehicles":               {},
	"books-sports-hobbies":   {},
	"kids":                   {},
	"pets":                   {},
	"services":               {},
	"smartphones":            {},
	"tablets":                {},
	"mobile-accessories":     {},
	"laptops-computers":      {},
	"tv-video-audio":         {},
	"cameras-accessories":    {},
	"home-appliances":        {},
	"sofas-chairs":           {},
	"beds-wardrobes":         {},
	"tables-dining":          {},
	"kitchen-dining":         {},
	"decor-lighting":         {},
	"garden-tools":           {},
	"fashion-men":            {},
	"fashion-women":          {},
	"fashion-accessories":    {},
	"cars":                   {},
	"motorcycles":            {},
	"bicycles":               {},
	"books":                  {},
	"sports-equipment":       {},
	"music-hobbies":          {},
	"baby-gear":              {},
	"toys-games":             {},
	"kids-fashion":           {},
	"pet-supplies":           {},
	"pet-adoption":           {},
	"home-services":          {},
	"repairs":                {},
	"tuition":                {},
}

func isCanonicalCategorySlug(slug string) bool {
	_, ok := canonicalCategorySlugs[slug]
	return ok
}

type Category struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ParentID *string `json:"parent_id"`
}

type CategoryTreeNode struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Slug     string             `json:"slug"`
	ParentID *string            `json:"parent_id"`
	Children []CategoryTreeNode `json:"children,omitempty"`
}

type CategoryStore struct {
	DB *sql.DB
}

func (s CategoryStore) List(ctx context.Context) ([]Category, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, slug, parent_id
		FROM categories
		ORDER BY (parent_id IS NOT NULL), name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {
		var c Category
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &parentID); err != nil {
			return nil, err
		}
		if parentID.Valid {
			value := parentID.String
			c.ParentID = &value
		}
		if !isCanonicalCategorySlug(c.Slug) {
			continue
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s CategoryStore) Tree(ctx context.Context) ([]CategoryTreeNode, error) {
	categories, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	return buildCategoryTree(categories), nil
}

func buildCategoryTree(categories []Category) []CategoryTreeNode {
	type mutableNode struct {
		CategoryTreeNode
		children []*mutableNode
	}

	nodesByID := make(map[string]*mutableNode, len(categories))
	for _, category := range categories {
		nodesByID[category.ID] = &mutableNode{
			CategoryTreeNode: CategoryTreeNode{
				ID:       category.ID,
				Name:     category.Name,
				Slug:     category.Slug,
				ParentID: category.ParentID,
			},
		}
	}

	var roots []*mutableNode
	for _, category := range categories {
		node := nodesByID[category.ID]
		if category.ParentID == nil {
			roots = append(roots, node)
			continue
		}
		parent, ok := nodesByID[*category.ParentID]
		if !ok {
			roots = append(roots, node)
			continue
		}
		parent.children = append(parent.children, node)
	}

	sortNodeList := func(nodes []*mutableNode) {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})
	}

	var freeze func(nodes []*mutableNode) []CategoryTreeNode
	freeze = func(nodes []*mutableNode) []CategoryTreeNode {
		sortNodeList(nodes)
		out := make([]CategoryTreeNode, 0, len(nodes))
		for _, node := range nodes {
			sortNodeList(node.children)
			frozen := node.CategoryTreeNode
			frozen.Children = freeze(node.children)
			out = append(out, frozen)
		}
		return out
	}

	return freeze(roots)
}
