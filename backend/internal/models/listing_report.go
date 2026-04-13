package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrListingReportOwnListing = errors.New("cannot report your own listing")

type ListingReportStore struct {
	DB *sql.DB
}

func (s ListingReportStore) Create(ctx context.Context, listingID, reporterID, reason string) error {
	var sellerID string
	err := s.DB.QueryRowContext(ctx, `
		SELECT seller_id
		FROM listings
		WHERE id = $1
		  AND deleted_at IS NULL
	`, listingID).Scan(&sellerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrListingNotFound
		}
		return err
	}

	if sellerID == reporterID {
		return ErrListingReportOwnListing
	}

	reason = strings.TrimSpace(reason)
	var reasonArg any
	if reason != "" {
		reasonArg = reason
	} else {
		reasonArg = nil
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO listing_reports (id, listing_id, reporter_id, reason)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (listing_id, reporter_id) DO NOTHING
	`, uuid.NewString(), listingID, reporterID, reasonArg)
	return err
}
