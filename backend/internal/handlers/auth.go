package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

type AuthHandler struct {
	Users                      models.UserStore
	TokenManager               utils.TokenManager
	TokenExpiryHours           int
	PasswordResetExpiryMinutes int
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type passwordResetRequest struct {
	Email string `json:"email"`
}

type passwordResetConfirmRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  publicUser `json:"user"`
}

type publicUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	City     string `json:"city,omitempty"`
	Bio      string `json:"bio,omitempty"`
	PhotoURL string `json:"photo_url,omitempty"`
}

type updateProfileRequest struct {
	FullName *string `json:"full_name"`
	City     *string `json:"city"`
	Bio      *string `json:"bio"`
	PhotoURL *string `json:"photo_url"`
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, password, and full_name are required"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to secure password"})
		return
	}

	user := models.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
	}

	createdUser, err := h.Users.Create(r.Context(), user)
	if err != nil {
		log.Printf("register create user failed: %v", err)
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate") || strings.Contains(lowerErr, "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		if strings.Contains(lowerErr, "relation \"users\" does not exist") {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "users table not found; run backend migration first"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	token, err := h.createToken(createdUser.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  toPublicUser(createdUser),
	})
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.Users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := h.createToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, User: toPublicUser(user)})
}

func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.Users.FindByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]publicUser{"user": toPublicUser(user)})
}

func (h AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	up := models.ProfileUpdate{
		FullName:        req.FullName,
		City:            req.City,
		Bio:             req.Bio,
		ProfileImageURL: req.PhotoURL,
	}

	user, err := h.Users.UpdateProfile(r.Context(), userID, up)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]publicUser{"user": toPublicUser(user)})
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Current auth is stateless JWT, so logout is client token disposal.
	// Keeping this endpoint enables forward compatibility with token revocation.
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req passwordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	user, err := h.Users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusOK, map[string]string{"message": "If an account exists, a password reset link has been sent"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}

	rawToken, err := generatePasswordResetToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.PasswordResetExpiryMinutes) * time.Minute)
	resetToken := models.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: passwordResetTokenHash(rawToken),
		ExpiresAt: expiresAt,
	}

	if err := h.Users.InvalidateActivePasswordResetTokensByUserID(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}
	if err := h.Users.CreatePasswordResetToken(r.Context(), resetToken); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}

	// TODO: Replace with email provider integration.
	log.Printf("password reset token generated for %s (expires %s): %s", req.Email, expiresAt.Format(time.RFC3339), rawToken)

	writeJSON(w, http.StatusOK, map[string]string{"message": "If an account exists, a password reset link has been sent"})
}

func (h AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req passwordResetConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Token == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token and new_password are required"})
		return
	}
	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new_password must be at least 8 characters"})
		return
	}

	userID, err := h.Users.ConsumePasswordResetToken(r.Context(), passwordResetTokenHash(req.Token))
	if err != nil {
		if errors.Is(err, models.ErrPasswordResetTokenInvalid) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired reset token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	passwordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	if err := h.Users.UpdatePasswordHashByID(r.Context(), userID, passwordHash); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired reset token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	if err := h.Users.InvalidateActivePasswordResetTokensByUserID(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}

func (h AuthHandler) createToken(userID string) (string, error) {
	expiresAt := time.Now().Add(time.Duration(h.TokenExpiryHours) * time.Hour)
	return h.TokenManager.Create(userID, expiresAt)
}

func generatePasswordResetToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func passwordResetTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func toPublicUser(user models.User) publicUser {
	return publicUser{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		City:     user.City,
		Bio:      user.Bio,
		PhotoURL: user.ProfileImageURL,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
