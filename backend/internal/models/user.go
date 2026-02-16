package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrUserNotFound = errors.New("user not found")
var ErrPasswordResetTokenInvalid = errors.New("password reset token is invalid or expired")
var ErrPasswordResetOTPInvalid = errors.New("password reset otp is invalid or expired")

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	FullName         string    `json:"full_name"`
	City             string    `json:"city,omitempty"`
	Bio              string    `json:"bio,omitempty"`
	ProfileImageURL  string    `json:"profile_image_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserStore struct {
	DB *sql.DB
}

type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetOTP struct {
	ID          string
	UserID      string
	OTPHash     string
	ExpiresAt   time.Time
	MaxAttempts int
}

func (s UserStore) Create(ctx context.Context, user User) (User, error) {
	query := `
		INSERT INTO users (id, email, password_hash, full_name)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := s.DB.QueryRowContext(ctx, query, user.ID, strings.ToLower(user.Email), user.PasswordHash, user.FullName).
		Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, err
	}

	user.Email = strings.ToLower(user.Email)
	return user, nil
}

func (s UserStore) FindByEmail(ctx context.Context, email string) (User, error) {
	query := `
		SELECT id, email, password_hash, full_name, city, bio, profile_image_url, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	var city, bio, profileImageURL sql.NullString
	err := s.DB.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&city,
		&bio,
		&profileImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	user.City = city.String
	user.Bio = bio.String
	user.ProfileImageURL = profileImageURL.String
	return user, nil
}

func (s UserStore) FindByID(ctx context.Context, id string) (User, error) {
	query := `
		SELECT id, email, password_hash, full_name, city, bio, profile_image_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	var city, bio, profileImageURL sql.NullString
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&city,
		&bio,
		&profileImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	user.City = city.String
	user.Bio = bio.String
	user.ProfileImageURL = profileImageURL.String
	return user, nil
}

func (s UserStore) InvalidateActivePasswordResetTokensByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE user_id = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
	`

	_, err := s.DB.ExecContext(ctx, query, userID)
	return err
}

func (s UserStore) CreatePasswordResetToken(ctx context.Context, token PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.DB.ExecContext(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (s UserStore) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (string, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	query := `
		SELECT id, user_id
		FROM password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`

	var tokenID string
	var userID string
	err = tx.QueryRowContext(ctx, query, tokenHash).Scan(&tokenID, &userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrPasswordResetTokenInvalid
		}
		return "", err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`, tokenID); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return userID, nil
}

// ProfileUpdate holds allowed profile fields. Only these may be updated via PATCH /users/me.
type ProfileUpdate struct {
	FullName        *string
	City            *string
	Bio             *string
	ProfileImageURL *string
}

func (s UserStore) UpdateProfile(ctx context.Context, userID string, u ProfileUpdate) (User, error) {
	user, err := s.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	if u.FullName != nil {
		trimmed := strings.TrimSpace(*u.FullName)
		if trimmed != "" {
			user.FullName = trimmed
		}
	}
	if u.City != nil {
		user.City = strings.TrimSpace(*u.City)
	}
	if u.Bio != nil {
		b := strings.TrimSpace(*u.Bio)
		if len(b) > 500 {
			b = b[:500]
		}
		user.Bio = b
	}
	if u.ProfileImageURL != nil {
		user.ProfileImageURL = strings.TrimSpace(*u.ProfileImageURL)
	}

	query := `
		UPDATE users
		SET full_name = $2, city = $3, bio = $4, profile_image_url = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING created_at, updated_at
	`

	err = s.DB.QueryRowContext(ctx, query,
		userID,
		user.FullName,
		nullIfEmpty(user.City),
		nullIfEmpty(user.Bio),
		nullIfEmpty(user.ProfileImageURL),
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (s UserStore) UpdatePasswordHashByID(ctx context.Context, userID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.DB.ExecContext(ctx, query, userID, passwordHash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s UserStore) InvalidateActivePasswordResetOTPsByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE password_reset_otps
		SET used_at = NOW()
		WHERE user_id = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
	`

	_, err := s.DB.ExecContext(ctx, query, userID)
	return err
}

func (s UserStore) CreatePasswordResetOTP(ctx context.Context, otp PasswordResetOTP) error {
	query := `
		INSERT INTO password_reset_otps (id, user_id, otp_hash, expires_at, max_attempts)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.DB.ExecContext(ctx, query, otp.ID, otp.UserID, otp.OTPHash, otp.ExpiresAt, otp.MaxAttempts)
	return err
}

func (s UserStore) ConsumePasswordResetOTP(ctx context.Context, userID, otpHash string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		SELECT id, otp_hash, attempt_count, max_attempts
		FROM password_reset_otps
		WHERE user_id = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`

	var id string
	var storedHash string
	var attemptCount int
	var maxAttempts int
	err = tx.QueryRowContext(ctx, query, userID).Scan(&id, &storedHash, &attemptCount, &maxAttempts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPasswordResetOTPInvalid
		}
		return err
	}

	if storedHash != otpHash {
		nextAttemptCount := attemptCount + 1
		if nextAttemptCount >= maxAttempts {
			if _, err := tx.ExecContext(ctx, `UPDATE password_reset_otps SET attempt_count = $2, used_at = NOW() WHERE id = $1`, id, nextAttemptCount); err != nil {
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx, `UPDATE password_reset_otps SET attempt_count = $2 WHERE id = $1`, id, nextAttemptCount); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		return ErrPasswordResetOTPInvalid
	}

	if _, err := tx.ExecContext(ctx, `UPDATE password_reset_otps SET used_at = NOW() WHERE id = $1`, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
