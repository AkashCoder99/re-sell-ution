package models

import (
	"context"
	"database/sql"
	"math"
	"time"
)

type FavoriteStore struct {
	DB *sql.DB
}

type Favorite struct {
	ListingID string    `json:"listing_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FavoritesPage struct {
	Favorites  []Favorite `json:"favorites"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}

func (s FavoriteStore) Add(ctx context.Context, userID, listingID string) error {
	var exists int
	err := s.DB.QueryRowContext(ctx, `
		SELECT 1
		FROM listings
		WHERE id = $1
		  AND deleted_at IS NULL
	`, listingID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrListingNotFound
		}
		return err
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO favorites (user_id, listing_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, listing_id) DO NOTHING
	`, userID, listingID)
	return err
}

func (s FavoriteStore) Remove(ctx context.Context, userID, listingID string) error {
	_, err := s.DB.ExecContext(ctx, `
		DELETE FROM favorites
		WHERE user_id = $1 AND listing_id = $2
	`, userID, listingID)
	return err
}

func (s FavoriteStore) Has(ctx context.Context, userID, listingID string) (bool, error) {
	var exists int
	err := s.DB.QueryRowContext(ctx, `
		SELECT 1
		FROM favorites f
		JOIN listings l ON l.id = f.listing_id
		WHERE f.user_id = $1
		  AND f.listing_id = $2
		  AND l.deleted_at IS NULL
	`, userID, listingID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s FavoriteStore) ListByUser(ctx context.Context, userID string, page, limit int) (FavoritesPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 5 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	var total int
	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM favorites f
		JOIN listings l ON l.id = f.listing_id
		WHERE f.user_id = $1
		  AND l.deleted_at IS NULL
	`, userID).Scan(&total); err != nil {
		return FavoritesPage{}, err
	}

	offset := (page - 1) * limit
	rows, err := s.DB.QueryContext(ctx, `
		SELECT f.listing_id, f.created_at
		FROM favorites f
		JOIN listings l ON l.id = f.listing_id
		WHERE f.user_id = $1
		  AND l.deleted_at IS NULL
		ORDER BY f.created_at DESC, f.listing_id ASC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return FavoritesPage{}, err
	}
	defer rows.Close()

	favorites := make([]Favorite, 0, limit)
	for rows.Next() {
		var item Favorite
		if err := rows.Scan(&item.ListingID, &item.CreatedAt); err != nil {
			return FavoritesPage{}, err
		}
		favorites = append(favorites, item)
	}
	if err := rows.Err(); err != nil {
		return FavoritesPage{}, err
	}

	totalPages := 1
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return FavoritesPage{
		Favorites:  favorites,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
