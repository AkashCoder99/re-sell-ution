package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrListingNotFound = errors.New("listing not found")

type Listing struct {
	ID           string         `json:"id"`
	SellerID     string         `json:"seller_id"`
	CategoryID   *string        `json:"category_id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Condition    string         `json:"condition"`
	Price        float64        `json:"price"`
	Currency     string         `json:"currency"`
	City         string         `json:"city"`
	State        string         `json:"state,omitempty"`
	Latitude     *float64       `json:"latitude,omitempty"`
	Longitude    *float64       `json:"longitude,omitempty"`
	Status       string         `json:"status"`
	ViewCount    int            `json:"view_count"`
	SoldToUserID *string        `json:"sold_to_user_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Images       []ListingImage `json:"images,omitempty"`
}

type ListingImage struct {
	ID        string    `json:"id"`
	ListingID string    `json:"listing_id"`
	ImageURL  string    `json:"image_url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type ListingCreate struct {
	CategoryID  *string
	Title       string
	Description string
	Condition   string
	Price       float64
	Currency    string
	City        string
	State       string
	Latitude    *float64
	Longitude   *float64
	ImageURLs   []string
}

type ListingPatch struct {
	Title          *string
	Description    *string
	Condition      *string
	Price          *float64
	Currency       *string
	City           *string
	State          *string
	CategoryIDSet  bool
	CategoryID     *string
	CoordinatesSet bool
	Latitude       *float64
	Longitude      *float64
}

type ListingStore struct {
	DB *sql.DB
}

func (s ListingStore) CategoryExists(ctx context.Context, id string) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT 1 FROM categories WHERE id = $1`, id).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s ListingStore) Create(ctx context.Context, sellerID, updatedBy string, in ListingCreate) (Listing, error) {
	if in.Currency == "" {
		in.Currency = "INR"
	}
	listingID := uuid.NewString()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Listing{}, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO listings (
			id, seller_id, category_id, title, description, condition,
			price, currency, city, state, latitude, longitude, status, updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'active', $13)
		RETURNING created_at, updated_at
	`
	var cat interface{}
	if in.CategoryID != nil && *in.CategoryID != "" {
		cat = *in.CategoryID
	} else {
		cat = nil
	}
	var state interface{}
	if strings.TrimSpace(in.State) != "" {
		state = strings.TrimSpace(in.State)
	} else {
		state = nil
	}
	var latitude interface{}
	if in.Latitude != nil {
		latitude = *in.Latitude
	} else {
		latitude = nil
	}
	var longitude interface{}
	if in.Longitude != nil {
		longitude = *in.Longitude
	} else {
		longitude = nil
	}

	var l Listing
	l.ID = listingID
	l.SellerID = sellerID
	if in.CategoryID != nil && *in.CategoryID != "" {
		cid := *in.CategoryID
		l.CategoryID = &cid
	}
	l.Title = in.Title
	l.Description = in.Description
	l.Condition = in.Condition
	l.Price = in.Price
	l.Currency = in.Currency
	l.City = in.City
	l.State = strings.TrimSpace(in.State)
	if in.Latitude != nil {
		value := *in.Latitude
		l.Latitude = &value
	}
	if in.Longitude != nil {
		value := *in.Longitude
		l.Longitude = &value
	}
	l.Status = "active"
	l.ViewCount = 0

	err = tx.QueryRowContext(ctx, query,
		listingID, sellerID, cat, in.Title, in.Description, in.Condition,
		in.Price, in.Currency, in.City, state, latitude, longitude, updatedBy,
	).Scan(&l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return Listing{}, err
	}

	for i, rawURL := range in.ImageURLs {
		url := strings.TrimSpace(rawURL)
		if url == "" {
			continue
		}
		imgID := uuid.NewString()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO listing_images (id, listing_id, image_url, position)
			VALUES ($1, $2, $3, $4)
		`, imgID, listingID, url, i)
		if err != nil {
			return Listing{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Listing{}, err
	}

	return s.loadListingWithImages(ctx, listingID)
}

func (s ListingStore) loadListingWithImages(ctx context.Context, listingID string) (Listing, error) {
	l, err := s.findByID(ctx, listingID)
	if err != nil {
		return Listing{}, err
	}
	imgs, err := s.listImagesByListingID(ctx, listingID)
	if err != nil {
		return Listing{}, err
	}
	l.Images = imgs
	return l, nil
}

func (s ListingStore) findByID(ctx context.Context, id string) (Listing, error) {
	query := `
		SELECT id, seller_id, category_id, title, description, condition,
		       price, currency, city, state, latitude, longitude, status, view_count,
		       sold_to_user_id, created_at, updated_at
		FROM listings
		WHERE id = $1
	`
	var l Listing
	var catID, st, soldTo sql.NullString
	var latitude, longitude sql.NullFloat64
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.SellerID, &catID, &l.Title, &l.Description, &l.Condition,
		&l.Price, &l.Currency, &l.City, &st, &latitude, &longitude, &l.Status, &l.ViewCount,
		&soldTo, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Listing{}, ErrListingNotFound
		}
		return Listing{}, err
	}
	if catID.Valid {
		v := catID.String
		l.CategoryID = &v
	}
	if st.Valid {
		l.State = st.String
	}
	if latitude.Valid {
		value := latitude.Float64
		l.Latitude = &value
	}
	if longitude.Valid {
		value := longitude.Float64
		l.Longitude = &value
	}
	if soldTo.Valid {
		v := soldTo.String
		l.SoldToUserID = &v
	}
	return l, nil
}

func (s ListingStore) listImagesByListingID(ctx context.Context, listingID string) ([]ListingImage, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, listing_id, image_url, position, created_at
		FROM listing_images
		WHERE listing_id = $1
		ORDER BY position ASC
	`, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ListingImage
	for rows.Next() {
		var img ListingImage
		if err := rows.Scan(&img.ID, &img.ListingID, &img.ImageURL, &img.Position, &img.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (s ListingStore) listingOwned(ctx context.Context, listingID, sellerID string) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `
		SELECT 1
		FROM listings
		WHERE id = $1 AND seller_id = $2 AND deleted_at IS NULL
	`, listingID, sellerID).Scan(&n)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s ListingStore) NextImagePosition(ctx context.Context, listingID, sellerID string) (int, error) {
	owned, err := s.listingOwned(ctx, listingID, sellerID)
	if err != nil {
		return 0, err
	}
	if !owned {
		return 0, ErrListingNotFound
	}

	// Next slot for this listing (avoids position collisions).
	var pos int
	err = s.DB.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position) + 1, 0)
		FROM listing_images
		WHERE listing_id = $1
	`, listingID).Scan(&pos)
	if err != nil {
		return 0, err
	}
	return pos, nil
}

// AddListingImage inserts a new listing_images row for an owned (non-deleted) listing.
func (s ListingStore) AddListingImage(ctx context.Context, listingID, sellerID, updatedBy, imageURL string, position int) (ListingImage, error) {
	if position < 0 {
		return ListingImage{}, errors.New("position must be non-negative")
	}

	owned, err := s.listingOwned(ctx, listingID, sellerID)
	if err != nil {
		return ListingImage{}, err
	}
	if !owned {
		return ListingImage{}, ErrListingNotFound
	}

	imgID := uuid.NewString()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return ListingImage{}, err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO listing_images (id, listing_id, image_url, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, listing_id, image_url, position, created_at
	`

	var img ListingImage
	if err := tx.QueryRowContext(ctx, insertQuery, imgID, listingID, imageURL, position).Scan(
		&img.ID,
		&img.ListingID,
		&img.ImageURL,
		&img.Position,
		&img.CreatedAt,
	); err != nil {
		return ListingImage{}, err
	}

	// Touch listing audit fields.
	updateQuery := `
		UPDATE listings
		SET updated_at = NOW(), updated_by = $3
		WHERE id = $1 AND seller_id = $2 AND deleted_at IS NULL
	`
	if _, err := tx.ExecContext(ctx, updateQuery, listingID, sellerID, updatedBy); err != nil {
		return ListingImage{}, err
	}

	if err := tx.Commit(); err != nil {
		return ListingImage{}, err
	}

	return img, nil
}

// FindOwned returns a non-deleted listing if it exists and belongs to sellerID.
func (s ListingStore) FindOwned(ctx context.Context, id, sellerID string) (Listing, error) {
	query := `
		SELECT id, seller_id, category_id, title, description, condition,
		       price, currency, city, state, latitude, longitude, status, view_count,
		       sold_to_user_id, created_at, updated_at
		FROM listings
		WHERE id = $1 AND seller_id = $2 AND deleted_at IS NULL
	`
	var l Listing
	var catID, st, soldTo sql.NullString
	var latitude, longitude sql.NullFloat64
	err := s.DB.QueryRowContext(ctx, query, id, sellerID).Scan(
		&l.ID, &l.SellerID, &catID, &l.Title, &l.Description, &l.Condition,
		&l.Price, &l.Currency, &l.City, &st, &latitude, &longitude, &l.Status, &l.ViewCount,
		&soldTo, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Listing{}, ErrListingNotFound
		}
		return Listing{}, err
	}
	if catID.Valid {
		v := catID.String
		l.CategoryID = &v
	}
	if st.Valid {
		l.State = st.String
	}
	if latitude.Valid {
		value := latitude.Float64
		l.Latitude = &value
	}
	if longitude.Valid {
		value := longitude.Float64
		l.Longitude = &value
	}
	if soldTo.Valid {
		v := soldTo.String
		l.SoldToUserID = &v
	}
	return l, nil
}

func (s ListingStore) Update(ctx context.Context, id, sellerID, updatedBy string, p ListingPatch) (Listing, error) {
	cur, err := s.FindOwned(ctx, id, sellerID)
	if err != nil {
		return Listing{}, err
	}

	if p.Title != nil {
		cur.Title = strings.TrimSpace(*p.Title)
	}
	if p.Description != nil {
		cur.Description = strings.TrimSpace(*p.Description)
	}
	if p.Condition != nil {
		cur.Condition = strings.TrimSpace(*p.Condition)
	}
	if p.Price != nil {
		cur.Price = *p.Price
	}
	if p.Currency != nil {
		cur.Currency = strings.TrimSpace(strings.ToUpper(*p.Currency))
	}
	if p.City != nil {
		cur.City = strings.TrimSpace(*p.City)
	}
	if p.State != nil {
		cur.State = strings.TrimSpace(*p.State)
	}
	if p.CategoryIDSet {
		if p.CategoryID == nil || strings.TrimSpace(*p.CategoryID) == "" {
			cur.CategoryID = nil
		} else {
			v := strings.TrimSpace(*p.CategoryID)
			cur.CategoryID = &v
		}
	}
	if p.CoordinatesSet {
		if p.Latitude == nil || p.Longitude == nil {
			cur.Latitude = nil
			cur.Longitude = nil
		} else {
			lat := *p.Latitude
			lng := *p.Longitude
			cur.Latitude = &lat
			cur.Longitude = &lng
		}
	}

	var cat interface{}
	if cur.CategoryID != nil && *cur.CategoryID != "" {
		cat = *cur.CategoryID
	} else {
		cat = nil
	}
	var state interface{}
	if cur.State != "" {
		state = cur.State
	} else {
		state = nil
	}
	var latitude interface{}
	if cur.Latitude != nil {
		latitude = *cur.Latitude
	} else {
		latitude = nil
	}
	var longitude interface{}
	if cur.Longitude != nil {
		longitude = *cur.Longitude
	} else {
		longitude = nil
	}

	query := `
		UPDATE listings
		SET title = $3, description = $4, condition = $5, price = $6,
		    currency = $7, city = $8, state = $9, category_id = $10,
		    latitude = $11, longitude = $12, updated_at = NOW(), updated_by = $13
		WHERE id = $1 AND seller_id = $2 AND deleted_at IS NULL
		RETURNING updated_at
	`
	var updatedAt time.Time
	err = s.DB.QueryRowContext(ctx, query,
		id, sellerID,
		cur.Title, cur.Description, cur.Condition, cur.Price,
		cur.Currency, cur.City, state, cat, latitude, longitude, updatedBy,
	).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Listing{}, ErrListingNotFound
		}
		return Listing{}, err
	}
	cur.UpdatedAt = updatedAt

	imgs, err := s.listImagesByListingID(ctx, id)
	if err != nil {
		return Listing{}, err
	}
	cur.Images = imgs
	return cur, nil
}

func (s ListingStore) SoftDelete(ctx context.Context, id, sellerID, updatedBy string) error {
	query := `
		UPDATE listings
		SET deleted_at = NOW(), status = 'deleted', updated_at = NOW(), updated_by = $3
		WHERE id = $1 AND seller_id = $2 AND deleted_at IS NULL
	`
	res, err := s.DB.ExecContext(ctx, query, id, sellerID, updatedBy)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrListingNotFound
	}
	return nil
}

func (s ListingStore) UpdateStatus(ctx context.Context, id, sellerID, updatedBy, status string, soldToUserID *string) (Listing, error) {
	cur, err := s.FindOwned(ctx, id, sellerID)
	if err != nil {
		return Listing{}, err
	}
	if status == "deleted" {
		return Listing{}, fmt.Errorf("invalid status")
	}

	var soldTo interface{}
	if soldToUserID != nil && *soldToUserID != "" {
		soldTo = *soldToUserID
	} else {
		soldTo = nil
	}
	if status != "sold" {
		soldTo = nil
	}

	query := `
		UPDATE listings
		SET status = $3, sold_to_user_id = $4, updated_at = NOW(), updated_by = $5
		WHERE id = $1 AND seller_id = $2 AND deleted_at IS NULL
		RETURNING updated_at
	`
	var updatedAt time.Time
	err = s.DB.QueryRowContext(ctx, query, id, sellerID, status, soldTo, updatedBy).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Listing{}, ErrListingNotFound
		}
		return Listing{}, err
	}
	cur.Status = status
	cur.SoldToUserID = nil
	if soldToUserID != nil && *soldToUserID != "" && status == "sold" {
		v := *soldToUserID
		cur.SoldToUserID = &v
	}
	cur.UpdatedAt = updatedAt

	imgs, err := s.listImagesByListingID(ctx, id)
	if err != nil {
		return Listing{}, err
	}
	cur.Images = imgs
	return cur, nil
}

type MyListingsPage struct {
	Listings   []Listing
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

type PublicListingsPage struct {
	Listings   []Listing
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

func (s ListingStore) ListPublic(ctx context.Context, city string, categoryID *string, page, limit int) (PublicListingsPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 12
	}
	if limit > 20 {
		limit = 20
	}

	city = strings.TrimSpace(city)
	if categoryID != nil {
		trimmed := strings.TrimSpace(*categoryID)
		categoryID = &trimmed
		if trimmed == "" {
			categoryID = nil
		}
	}

	where := []string{"status = 'active'", "deleted_at IS NULL"}
	args := []any{}
	if city != "" {
		args = append(args, city)
		where = append(where, fmt.Sprintf("lower(city) = lower($%d)", len(args)))
	}
	if categoryID != nil {
		args = append(args, *categoryID)
		where = append(where, fmt.Sprintf("category_id = $%d", len(args)))
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM listings
		WHERE %s
	`, strings.Join(where, " AND "))

	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return PublicListingsPage{}, err
	}

	offset := (page - 1) * limit
	listArgs := append(append([]any{}, args...), limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	query := fmt.Sprintf(`
		SELECT id, seller_id, category_id, title, description, condition,
		       price, currency, city, state, latitude, longitude, status, view_count,
		       sold_to_user_id, created_at, updated_at
		FROM listings
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), limitPos, offsetPos)

	rows, err := s.DB.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return PublicListingsPage{}, err
	}
	defer rows.Close()

	listings := make([]Listing, 0)
	for rows.Next() {
		var l Listing
		var catID, st, soldTo sql.NullString
		var latitude, longitude sql.NullFloat64
		if err := rows.Scan(
			&l.ID, &l.SellerID, &catID, &l.Title, &l.Description, &l.Condition,
			&l.Price, &l.Currency, &l.City, &st, &latitude, &longitude, &l.Status, &l.ViewCount,
			&soldTo, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return PublicListingsPage{}, err
		}
		if catID.Valid {
			v := catID.String
			l.CategoryID = &v
		}
		if st.Valid {
			l.State = st.String
		}
		if latitude.Valid {
			value := latitude.Float64
			l.Latitude = &value
		}
		if longitude.Valid {
			value := longitude.Float64
			l.Longitude = &value
		}
		if soldTo.Valid {
			v := soldTo.String
			l.SoldToUserID = &v
		}
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		return PublicListingsPage{}, err
	}

	for i := range listings {
		imgs, err := s.listImagesByListingID(ctx, listings[i].ID)
		if err != nil {
			return PublicListingsPage{}, err
		}
		listings[i].Images = imgs
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return PublicListingsPage{
		Listings:   listings,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s ListingStore) ListBySeller(ctx context.Context, sellerID, statusFilter string, page, limit int) (MyListingsPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 5 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	if statusFilter == "draft" {
		return MyListingsPage{
			Listings:   []Listing{},
			Total:      0,
			Page:       page,
			Limit:      limit,
			TotalPages: 1,
		}, nil
	}

	var countQuery string
	var args []interface{}
	args = append(args, sellerID)
	argN := 2
	statusClause := ""
	if statusFilter != "" && statusFilter != "all" {
		statusClause = fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, statusFilter)
		argN++
	}

	countQuery = fmt.Sprintf(`
		SELECT COUNT(*) FROM listings
		WHERE seller_id = $1 AND deleted_at IS NULL%s
	`, statusClause)

	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return MyListingsPage{}, err
	}

	offset := (page - 1) * limit
	listArgs := []interface{}{sellerID}
	listArgN := 2
	listStatusClause := ""
	if statusFilter != "" && statusFilter != "all" {
		listStatusClause = fmt.Sprintf(" AND status = $%d", listArgN)
		listArgs = append(listArgs, statusFilter)
		listArgN++
	}
	listArgs = append(listArgs, limit, offset)
	limitPos := listArgN
	offsetPos := listArgN + 1

	query := fmt.Sprintf(`
		SELECT id, seller_id, category_id, title, description, condition,
		       price, currency, city, state, latitude, longitude, status, view_count,
		       sold_to_user_id, created_at, updated_at
		FROM listings
		WHERE seller_id = $1 AND deleted_at IS NULL%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, listStatusClause, limitPos, offsetPos)

	rows, err := s.DB.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return MyListingsPage{}, err
	}
	defer rows.Close()

	var listings []Listing
	for rows.Next() {
		var l Listing
		var catID, st, soldTo sql.NullString
		var latitude, longitude sql.NullFloat64
		if err := rows.Scan(
			&l.ID, &l.SellerID, &catID, &l.Title, &l.Description, &l.Condition,
			&l.Price, &l.Currency, &l.City, &st, &latitude, &longitude, &l.Status, &l.ViewCount,
			&soldTo, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return MyListingsPage{}, err
		}
		if catID.Valid {
			v := catID.String
			l.CategoryID = &v
		}
		if st.Valid {
			l.State = st.String
		}
		if latitude.Valid {
			value := latitude.Float64
			l.Latitude = &value
		}
		if longitude.Valid {
			value := longitude.Float64
			l.Longitude = &value
		}
		if soldTo.Valid {
			v := soldTo.String
			l.SoldToUserID = &v
		}
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		return MyListingsPage{}, err
	}

	for i := range listings {
		imgs, err := s.listImagesByListingID(ctx, listings[i].ID)
		if err != nil {
			return MyListingsPage{}, err
		}
		listings[i].Images = imgs
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return MyListingsPage{
		Listings:   listings,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
