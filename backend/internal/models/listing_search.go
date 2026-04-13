package models

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

const listingSearchVectorSQL = `
	setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
	setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
`

type ListingSearchParams struct {
	Query      string
	TSQuery    string
	City       string
	CategoryID *string
	Condition  *string
	MinPrice   *float64
	MaxPrice   *float64
	Sort       string
	Latitude   *float64
	Longitude  *float64
	RadiusKM   *float64
	Page       int
	Limit      int
}

type ListingSearchPage struct {
	Listings   []Listing
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

func (s ListingStore) Search(ctx context.Context, params ListingSearchParams) (ListingSearchPage, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 12
	}
	if params.Limit > 20 {
		params.Limit = 20
	}

	whereClause, args, distanceExpr := buildListingSearchFilters(params)

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM listings l
		WHERE %s
	`, whereClause)

	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return ListingSearchPage{}, err
	}

	offset := (params.Page - 1) * params.Limit
	listArgs := append(append([]any{}, args...), params.Limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	orderClause := buildListingSearchOrderClause(params.Sort, distanceExpr)

	query := fmt.Sprintf(`
		SELECT id, seller_id, category_id, title, description, condition,
		       price, currency, city, state, latitude, longitude, status, view_count,
		       sold_to_user_id, created_at, updated_at
		FROM listings l
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, limitPos, offsetPos)

	rows, err := s.DB.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return ListingSearchPage{}, err
	}
	defer rows.Close()

	var listings []Listing
	for rows.Next() {
		var listing Listing
		var categoryID, state, soldTo sql.NullString
		var latitude, longitude sql.NullFloat64
		if err := rows.Scan(
			&listing.ID, &listing.SellerID, &categoryID, &listing.Title, &listing.Description, &listing.Condition,
			&listing.Price, &listing.Currency, &listing.City, &state, &latitude, &longitude, &listing.Status, &listing.ViewCount,
			&soldTo, &listing.CreatedAt, &listing.UpdatedAt,
		); err != nil {
			return ListingSearchPage{}, err
		}
		if categoryID.Valid {
			value := categoryID.String
			listing.CategoryID = &value
		}
		if state.Valid {
			listing.State = state.String
		}
		if latitude.Valid {
			value := latitude.Float64
			listing.Latitude = &value
		}
		if longitude.Valid {
			value := longitude.Float64
			listing.Longitude = &value
		}
		if soldTo.Valid {
			value := soldTo.String
			listing.SoldToUserID = &value
		}
		listings = append(listings, listing)
	}
	if err := rows.Err(); err != nil {
		return ListingSearchPage{}, err
	}

	for i := range listings {
		images, err := s.listImagesByListingID(ctx, listings[i].ID)
		if err != nil {
			return ListingSearchPage{}, err
		}
		listings[i].Images = images
	}

	totalPages := 1
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.Limit)))
	}

	return ListingSearchPage{
		Listings:   listings,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

func buildListingSearchFilters(params ListingSearchParams) (string, []any, string) {
	args := []any{params.TSQuery}
	where := []string{
		"l.deleted_at IS NULL",
		"l.status = 'active'",
		fmt.Sprintf("(%s) @@ to_tsquery('simple', $1)", listingSearchVectorSQL),
	}
	distanceExpr := "NULL::double precision"

	if strings.TrimSpace(params.City) != "" {
		args = append(args, strings.TrimSpace(params.City))
		where = append(where, fmt.Sprintf("lower(l.city) = lower($%d)", len(args)))
	}

	if params.CategoryID != nil && strings.TrimSpace(*params.CategoryID) != "" {
		args = append(args, strings.TrimSpace(*params.CategoryID))
		where = append(where, fmt.Sprintf("l.category_id = $%d", len(args)))
	}

	if params.Condition != nil && strings.TrimSpace(*params.Condition) != "" {
		args = append(args, strings.TrimSpace(*params.Condition))
		where = append(where, fmt.Sprintf("l.condition = $%d", len(args)))
	}

	if params.MinPrice != nil {
		args = append(args, *params.MinPrice)
		where = append(where, fmt.Sprintf("l.price >= $%d", len(args)))
	}
	if params.MaxPrice != nil {
		args = append(args, *params.MaxPrice)
		where = append(where, fmt.Sprintf("l.price <= $%d", len(args)))
	}

	if params.Latitude != nil && params.Longitude != nil && params.RadiusKM != nil {
		latMin, latMax, lngMin, lngMax := geoBoundingBox(*params.Latitude, *params.Longitude, *params.RadiusKM)

		args = append(args, latMin, latMax, lngMin, lngMax)
		latMinPos := len(args) - 3
		latMaxPos := len(args) - 2
		lngMinPos := len(args) - 1
		lngMaxPos := len(args)

		where = append(where,
			"l.latitude IS NOT NULL",
			"l.longitude IS NOT NULL",
			fmt.Sprintf("l.latitude BETWEEN $%d AND $%d", latMinPos, latMaxPos),
			fmt.Sprintf("l.longitude BETWEEN $%d AND $%d", lngMinPos, lngMaxPos),
		)

		args = append(args, *params.Latitude, *params.Longitude, *params.RadiusKM)
		latPos := len(args) - 2
		lngPos := len(args) - 1
		radiusPos := len(args)

		distanceExpr = geoDistanceSQL(latPos, lngPos)
		where = append(where, fmt.Sprintf("%s <= $%d", distanceExpr, radiusPos))
	}

	return strings.Join(where, " AND "), args, distanceExpr
}

func buildListingSearchOrderClause(sortBy, distanceExpr string) string {
	switch sortBy {
	case "created_at_asc":
		return "l.created_at ASC, l.id ASC"
	case "created_at_desc":
		return "l.created_at DESC, l.id ASC"
	case "price_asc":
		return "l.price ASC, l.created_at DESC, l.id ASC"
	case "price_desc":
		return "l.price DESC, l.created_at DESC, l.id ASC"
	default:
		distanceOrder := ""
		if distanceExpr != "NULL::double precision" {
			distanceOrder = distanceExpr + " ASC NULLS LAST, "
		}
		return fmt.Sprintf("ts_rank_cd((%s), to_tsquery('simple', $1)) DESC, %s l.created_at DESC, l.id ASC", listingSearchVectorSQL, distanceOrder)
	}
}

func geoBoundingBox(latitude, longitude, radiusKM float64) (float64, float64, float64, float64) {
	latDelta := radiusKM / 111.0
	cosLat := math.Abs(math.Cos(latitude * math.Pi / 180))
	if cosLat < 0.01 {
		cosLat = 0.01
	}
	lngDelta := radiusKM / (111.0 * cosLat)
	return latitude - latDelta, latitude + latDelta, longitude - lngDelta, longitude + lngDelta
}

func geoDistanceSQL(latPos, lngPos int) string {
	return fmt.Sprintf(`
		6371.0 * ACOS(
			LEAST(
				1.0,
				GREATEST(
					-1.0,
					COS(RADIANS($%d)) * COS(RADIANS(l.latitude)) * COS(RADIANS(l.longitude) - RADIANS($%d)) +
					SIN(RADIANS($%d)) * SIN(RADIANS(l.latitude))
				)
			)
		)
	`, latPos, lngPos, latPos)
}
