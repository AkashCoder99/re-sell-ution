package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

type AuthHandler struct {
	Users                      models.UserStore
	TokenManager               utils.TokenManager
	EmailSender                utils.EmailSender
	TokenExpiryHours             int
	PasswordResetExpiryMinutes   int
	PasswordResetCooldownMinutes int
	PasswordResetOTPDigits       int
	PasswordResetMaxAttempts     int
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
	Email       string `json:"email"`
	OTP         string `json:"otp"`
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

const (
	maxEmailLength    = 254
	minPasswordLength = 8
	maxPasswordLength = 72
	minFullNameLength = 2
	maxFullNameLength = 100
	maxCityLength     = 100
	maxBioLength      = 500
	maxPhotoURLLength = 2048
)

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
	if err := validateEmail(req.Email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := validateFullName(req.FullName); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := validatePassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
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
	log.Printf("auth.register.success email=%s user_id=%s", createdUser.Email, createdUser.ID)

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
	if err := validateEmail(req.Email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(req.Password) > maxPasswordLength {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("password must not exceed %d characters", maxPasswordLength)})
		return
	}

	user, err := h.Users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			log.Printf("auth.login.failed email=%s reason=user_not_found", req.Email)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		log.Printf("auth.login.failed email=%s reason=fetch_error err=%v", req.Email, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		log.Printf("auth.login.failed email=%s reason=invalid_password", req.Email)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := h.createToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, User: toPublicUser(user)})
	log.Printf("auth.login.success email=%s user_id=%s", user.Email, user.ID)
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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if err := validateProfileUpdate(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	up := models.ProfileUpdate{
		FullName:        req.FullName,
		City:            req.City,
		Bio:             req.Bio,
		ProfileImageURL: req.PhotoURL,
	}

	user, err := h.Users.UpdateProfile(r.Context(), userID, userID, up)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]publicUser{"user": toPublicUser(user)})
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Current auth is stateless JWT, so logout is client token disposal.
	// Keeping this endpoint enables forward compatibility with token revocation.
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
	log.Printf("auth.logout.success user_id=%s", userID)
}

func (h AuthHandler) DeactivateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.Users.DeactivateByID(r.Context(), userID, userID); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
			return
		}
		log.Printf("auth.deactivate failed user_id=%s err=%v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to deactivate account"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "account deactivated successfully"})
	log.Printf("auth.deactivate.success user_id=%s", userID)
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
			log.Printf("auth.password_reset.request ignored email=%s reason=user_not_found", req.Email)
			writeJSON(w, http.StatusOK, map[string]string{"message": "If an account exists, an OTP has been sent to the registered email"})
			return
		}
		log.Printf("auth.password_reset.request failed email=%s err=%v", req.Email, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}

	cooldown := h.passwordResetCooldownMinutes()
	if cooldown > 0 {
		lastRequestAt, recent, err := h.Users.GetLastPasswordResetRequestTime(r.Context(), user.ID, cooldown)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
			return
		}
		if recent {
			cooldownEndsAt := lastRequestAt.Add(time.Duration(cooldown) * time.Minute)
			minutesLeft := int(time.Until(cooldownEndsAt).Minutes())
			if minutesLeft < 1 {
				minutesLeft = 1
			}
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"error":                fmt.Sprintf("Please wait %d more minute(s) before requesting another reset code", minutesLeft),
				"retry_after_minutes": minutesLeft,
			})
			return
		}
	}

	otp, err := generateNumericOTP(h.passwordResetOTPDigits())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.PasswordResetExpiryMinutes) * time.Minute)
	resetOTP := models.PasswordResetOTP{
		ID:          uuid.NewString(),
		UserID:      user.ID,
		OTPHash:     passwordResetOTPHash(otp),
		ExpiresAt:   expiresAt,
		MaxAttempts: h.passwordResetMaxAttempts(),
	}

	if err := h.Users.InvalidateActivePasswordResetOTPsByUserID(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}
	if err := h.Users.CreatePasswordResetOTP(r.Context(), resetOTP); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
		return
	}

	subject := "ReSellution Password Reset OTP"
	body := fmt.Sprintf(
		"Your ReSellution password reset OTP is %s.\n\nThis code expires in %d minutes.\nIf you did not request this, please ignore this email.",
		otp,
		h.PasswordResetExpiryMinutes,
	)

	if h.EmailSender != nil {
		if err := h.EmailSender.Send(req.Email, subject, body); err != nil {
			log.Printf("failed to send password reset OTP email to %s: %v", req.Email, err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process reset request"})
			return
		}
	} else {
		log.Printf("SMTP not configured. OTP for %s (expires %s): %s", req.Email, expiresAt.Format(time.RFC3339), otp)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "If an account exists, an OTP has been sent to the registered email"})
	log.Printf("auth.password_reset.request success email=%s user_id=%s", req.Email, user.ID)
}

func (h AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req passwordResetConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.OTP = strings.TrimSpace(req.OTP)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, otp, and new_password are required"})
		return
	}
	if err := validateEmail(req.Email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := validatePassword(req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": strings.Replace(err.Error(), "password", "new_password", 1)})
		return
	}

	user, err := h.Users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			log.Printf("auth.password_reset.confirm failed email=%s reason=user_not_found", req.Email)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired otp"})
			return
		}
		log.Printf("auth.password_reset.confirm failed email=%s reason=fetch_error err=%v", req.Email, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	if err := h.Users.ConsumePasswordResetOTP(r.Context(), user.ID, passwordResetOTPHash(req.OTP)); err != nil {
		if errors.Is(err, models.ErrPasswordResetOTPInvalid) {
			log.Printf("auth.password_reset.confirm failed email=%s user_id=%s reason=invalid_otp", req.Email, user.ID)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired otp"})
			return
		}
		log.Printf("auth.password_reset.confirm failed email=%s user_id=%s reason=consume_error err=%v", req.Email, user.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	passwordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	if err := h.Users.UpdatePasswordHashByID(r.Context(), user.ID, passwordHash, user.ID); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired otp"})
			return
		}
		log.Printf("auth.password_reset.confirm failed email=%s user_id=%s reason=update_error err=%v", req.Email, user.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	if err := h.Users.InvalidateActivePasswordResetOTPsByUserID(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset password"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
	log.Printf("auth.password_reset.confirm success email=%s user_id=%s", req.Email, user.ID)
}

func (h AuthHandler) createToken(userID string) (string, error) {
	expiresAt := time.Now().Add(time.Duration(h.TokenExpiryHours) * time.Hour)
	return h.TokenManager.Create(userID, expiresAt)
}

func generateNumericOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("otp length must be positive")
	}

	var builder strings.Builder
	builder.Grow(length)
	limit := big.NewInt(10)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + n.Int64()))
	}

	return builder.String(), nil
}

func passwordResetOTPHash(otp string) string {
	sum := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(sum[:])
}

func (h AuthHandler) passwordResetCooldownMinutes() int {
	if h.PasswordResetCooldownMinutes < 0 {
		return 0
	}
	return h.PasswordResetCooldownMinutes
}

func (h AuthHandler) passwordResetOTPDigits() int {
	if h.PasswordResetOTPDigits <= 0 {
		return 6
	}
	return h.PasswordResetOTPDigits
}

func (h AuthHandler) passwordResetMaxAttempts() int {
	if h.PasswordResetMaxAttempts <= 0 {
		return 5
	}
	return h.PasswordResetMaxAttempts
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

func validateEmail(email string) error {
	if len(email) > maxEmailLength {
		return fmt.Errorf("email must not exceed %d characters", maxEmailLength)
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || strings.Count(email, "@") != 1 {
		return errors.New("email format is invalid")
	}

	return nil
}

func validateFullName(fullName string) error {
	if len(fullName) < minFullNameLength {
		return fmt.Errorf("full_name must be at least %d characters", minFullNameLength)
	}
	if len(fullName) > maxFullNameLength {
		return fmt.Errorf("full_name must not exceed %d characters", maxFullNameLength)
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	if len(password) > maxPasswordLength {
		return fmt.Errorf("password must not exceed %d characters", maxPasswordLength)
	}

	var hasLetter bool
	var hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("password must include at least one letter and one number")
	}

	return nil
}

func validateProfileUpdate(req *updateProfileRequest) error {
	if req.FullName == nil && req.City == nil && req.Bio == nil && req.PhotoURL == nil {
		return errors.New("at least one profile field is required")
	}

	if req.FullName != nil {
		trimmed := strings.TrimSpace(*req.FullName)
		if err := validateFullName(trimmed); err != nil {
			return err
		}
		*req.FullName = trimmed
	}

	if req.City != nil {
		trimmed := strings.TrimSpace(*req.City)
		if utf8.RuneCountInString(trimmed) > maxCityLength {
			return fmt.Errorf("city must not exceed %d characters", maxCityLength)
		}
		*req.City = trimmed
	}

	if req.Bio != nil {
		trimmed := strings.TrimSpace(*req.Bio)
		if utf8.RuneCountInString(trimmed) > maxBioLength {
			return fmt.Errorf("bio must not exceed %d characters", maxBioLength)
		}
		*req.Bio = trimmed
	}

	if req.PhotoURL != nil {
		trimmed := strings.TrimSpace(*req.PhotoURL)
		if trimmed != "" {
			if len(trimmed) > maxPhotoURLLength {
				return fmt.Errorf("photo_url must not exceed %d characters", maxPhotoURLLength)
			}
			parsed, err := url.ParseRequestURI(trimmed)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return errors.New("photo_url must be a valid absolute URL")
			}
			if parsed.Scheme != "http" && parsed.Scheme != "https" {
				return errors.New("photo_url must start with http:// or https://")
			}
		}
		*req.PhotoURL = trimmed
	}

	return nil
}
