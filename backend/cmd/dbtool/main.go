package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"resellution/backend/internal/db"
)

var migrationFiles = []string{
	"migrations/0001_init.sql",
	"migrations/0002_password_reset_tokens.sql",
	"migrations/0003_password_reset_otps.sql",
	"migrations/0004_add_user_bio.sql",
	"migrations/0005_b5_constraints_indexes.sql",
	"migrations/0007_soft_delete_audit.sql",
	"migrations/0006_notifications.sql",
	"migrations/0008_updated_by.sql",
	"migrations/0009_listing_sold_to.sql",
	"migrations/0010_search_geo.sql",
	"migrations/0011_chat_model_indexes.sql",
	"migrations/0012_listing_reports.sql",
	"migrations/0013_moderation_core.sql",
	"migrations/0014_moderation_integrity_retention_backfill.sql",
	"migrations/0013_listing_draft_status.sql",
	"migrations/0015_admin_users.sql",
}

var seedFiles = []string{
	"seeds/001_categories.sql",
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/dbtool <migrate|seed|setup|verify-search-plans>")
	}

	loadDotEnv(".env")
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	switch os.Args[1] {
	case "migrate":
		err = runSQLFiles(ctx, conn, migrationFiles)
	case "seed":
		err = runSQLFiles(ctx, conn, seedFiles)
	case "setup":
		if err = runSQLFiles(ctx, conn, migrationFiles); err == nil {
			err = runSQLFiles(ctx, conn, seedFiles)
		}
	case "verify-search-plans":
		err = verifySearchPlans(ctx, conn)
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runSQLFiles(ctx context.Context, conn *sql.DB, paths []string) error {
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		for _, statement := range splitSQLStatements(string(content)) {
			if _, err := conn.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}

		fmt.Println("applied", path)
	}
	return nil
}

func verifySearchPlans(ctx context.Context, conn *sql.DB) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name)
		VALUES (
			'00000000-0000-0000-0000-000000999001',
			'search-plan@example.com',
			'verification-only',
			'Search Verification'
		)
		ON CONFLICT (email) DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO listings (
			id,
			seller_id,
			title,
			description,
			condition,
			price,
			currency,
			city,
			state,
			latitude,
			longitude,
			status,
			updated_by
		)
		SELECT
			('00000000-0000-0000-0000-' || LPAD(g::text, 12, '0'))::uuid,
			'00000000-0000-0000-0000-000000999001',
			CASE
				WHEN g % 17 = 0 THEN 'mountain bike helmet ' || g
				WHEN g % 13 = 0 THEN 'iphone charger ' || g
				WHEN g % 11 = 0 THEN 'study desk ' || g
				ELSE 'marketplace listing ' || g
			END,
			CASE
				WHEN g % 17 = 0 THEN 'Protective bike helmet for daily rides and campus commutes'
				WHEN g % 13 = 0 THEN 'Fast charging adapter and cable bundle'
				WHEN g % 11 = 0 THEN 'Wooden desk with storage shelves'
				ELSE 'General purpose listing used to exercise search indexes'
			END,
			CASE
				WHEN g % 5 = 0 THEN 'like_new'
				WHEN g % 5 = 1 THEN 'good'
				WHEN g % 5 = 2 THEN 'fair'
				WHEN g % 5 = 3 THEN 'new'
				ELSE 'poor'
			END,
			10 + g,
			'INR',
			CASE
				WHEN g % 3 = 0 THEN 'Gainesville'
				WHEN g % 3 = 1 THEN 'Orlando'
				ELSE 'Tampa'
			END,
			'FL',
			CASE
				WHEN g % 3 = 0 THEN 29.6516 + ((g % 25) * 0.001)
				WHEN g % 3 = 1 THEN 28.5383 + ((g % 25) * 0.001)
				ELSE 27.9506 + ((g % 25) * 0.001)
			END,
			CASE
				WHEN g % 3 = 0 THEN -82.3248 + ((g % 25) * 0.001)
				WHEN g % 3 = 1 THEN -81.3792 + ((g % 25) * 0.001)
				ELSE -82.4572 + ((g % 25) * 0.001)
			END,
			'active',
			'00000000-0000-0000-0000-000000999001'
		FROM generate_series(1, 3000) AS g
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `ANALYZE listings`); err != nil {
		return err
	}

	plans := []struct {
		name  string
		query string
	}{
		{
			name: "FTS only plan",
			query: `
				SELECT l.id
				FROM listings l
				WHERE l.deleted_at IS NULL
				  AND l.status = 'active'
				  AND (
				      setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
				      setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
				  ) @@ to_tsquery('simple', 'bike:* & helmet:*')
				ORDER BY ts_rank_cd(
				    (
				        setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
				        setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
				    ),
				    to_tsquery('simple', 'bike:* & helmet:*')
				) DESC,
				l.created_at DESC
				LIMIT 20
			`,
		},
		{
			name: "FTS + city plan",
			query: `
				SELECT l.id
				FROM listings l
				WHERE l.deleted_at IS NULL
				  AND l.status = 'active'
				  AND lower(l.city) = lower('Gainesville')
				  AND (
				      setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
				      setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
				  ) @@ to_tsquery('simple', 'bike:* & helmet:*')
				ORDER BY ts_rank_cd(
				    (
				        setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
				        setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
				    ),
				    to_tsquery('simple', 'bike:* & helmet:*')
				) DESC,
				l.created_at DESC
				LIMIT 20
			`,
		},
		{
			name: "FTS + city + radius plan",
			query: `
				SELECT l.id
				FROM listings l
				WHERE l.deleted_at IS NULL
				  AND l.status = 'active'
				  AND lower(l.city) = lower('Gainesville')
				  AND l.latitude BETWEEN 29.6516 - (25.0 / 111.0) AND 29.6516 + (25.0 / 111.0)
				  AND l.longitude BETWEEN -82.3248 - (25.0 / (111.0 * GREATEST(COS(RADIANS(29.6516)), 0.01)))
				                        AND -82.3248 + (25.0 / (111.0 * GREATEST(COS(RADIANS(29.6516)), 0.01)))
				  AND (
				      6371.0 * ACOS(
				          LEAST(
				              1.0,
				              GREATEST(
				                  -1.0,
				                  COS(RADIANS(29.6516)) * COS(RADIANS(l.latitude)) * COS(RADIANS(l.longitude) - RADIANS(-82.3248)) +
				                  SIN(RADIANS(29.6516)) * SIN(RADIANS(l.latitude))
				              )
				          )
				      )
				  ) <= 25.0
				  AND (
				      setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
				      setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
				  ) @@ to_tsquery('simple', 'bike:* & helmet:*')
				ORDER BY ts_rank_cd(
				    (
				        setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
				        setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
				    ),
				    to_tsquery('simple', 'bike:* & helmet:*')
				) DESC,
				l.created_at DESC
				LIMIT 20
			`,
		},
	}

	for _, plan := range plans {
		fmt.Println(plan.name)
		rows, err := tx.QueryContext(ctx, "EXPLAIN (ANALYZE, COSTS OFF, BUFFERS) "+plan.query)
		if err != nil {
			return err
		}

		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				rows.Close()
				return err
			}
			fmt.Println(line)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}

	return tx.Rollback()
}

func splitSQLStatements(input string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	inLineComment := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			current.WriteByte(ch)
			continue
		}

		if !inString && ch == '-' && i+1 < len(input) && input[i+1] == '-' {
			inLineComment = true
			current.WriteByte(ch)
			continue
		}

		if ch == '\'' {
			inString = !inString
		}

		if ch == ';' && !inString {
			statement := strings.TrimSpace(current.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	statement := strings.TrimSpace(current.String())
	if statement != "" {
		statements = append(statements, statement)
	}

	return statements
}

func loadDotEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil && !errors.Is(err, os.ErrPermission) {
			continue
		}
	}
}
