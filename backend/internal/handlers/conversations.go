package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
	"resellution/backend/internal/utils"
)

const (
	maxConversationSearchLen         = 200
	maxConversationDetailMessagePage = 50
	maxConversationMessageLen        = 2000
)

type conversationStore interface {
	CreateOrGet(ctx context.Context, listingID, buyerID string) (models.Conversation, error)
	ListByUser(ctx context.Context, userID, query string, page, limit int) (models.ConversationsPage, error)
	GetByID(ctx context.Context, conversationID, userID string, page, limit int) (models.Conversation, error)
	ListMessages(ctx context.Context, conversationID, userID string, page, limit int) (models.MessagesPage, error)
	AddMessage(ctx context.Context, conversationID, userID, text string) (models.Message, *models.MessageNotificationDelivery, error)
	MarkRead(ctx context.Context, conversationID, userID string) error
}

type ConversationHandler struct {
	Conversations conversationStore
	EmailSender   utils.EmailSender
}

type createConversationRequest struct {
	BuyerID      string  `json:"buyer_id"`
	SellerID     string  `json:"seller_id"`
	SellerName   string  `json:"seller_name"`
	ListingID    string  `json:"listing_id"`
	ListingTitle string  `json:"listing_title"`
	ListingPrice float64 `json:"listing_price"`
	ListingCity  string  `json:"listing_city"`
}

type sendMessageRequest struct {
	SenderID string `json:"sender_id"`
	Text     string `json:"text"`
}

func (h ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createConversationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.ListingID = strings.TrimSpace(req.ListingID)
	if req.ListingID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "listing_id is required"})
		return
	}
	if _, err := uuid.Parse(req.ListingID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	if claimedBuyerID := strings.TrimSpace(req.BuyerID); claimedBuyerID != "" && claimedBuyerID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	conversation, err := h.Conversations.CreateOrGet(r.Context(), req.ListingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrListingNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "listing not found"})
			return
		case errors.Is(err, models.ErrConversationSelfNotAllowed):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot start a conversation on your own listing"})
			return
		case errors.Is(err, models.ErrConversationForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		default:
			observability.Error(r.Context(), "chat.conversations.create.failed", map[string]any{
				"user_id":    userID,
				"listing_id": req.ListingID,
				"error":      err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start conversation"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"conversation": conversation})
}

func (h ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(query) > maxConversationSearchLen {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "q is too long"})
		return
	}

	page, limit, err := parsePaginationParams(r, 1, 8)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	res, err := h.Conversations.ListByUser(r.Context(), userID, query, page, limit)
	if err != nil {
		observability.Error(r.Context(), "chat.conversations.list.failed", map[string]any{
			"user_id": userID,
			"query":   query,
			"error":   err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load conversations"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversations": res.Conversations,
		"total":         res.Total,
		"page":          res.Page,
		"limit":         res.Limit,
		"total_pages":   res.TotalPages,
	})
}

func (h ConversationHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	conversationID, valid := validateConversationID(w, r)
	if !valid {
		return
	}

	conversation, err := h.Conversations.GetByID(r.Context(), conversationID, userID, 1, maxConversationDetailMessagePage)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrConversationNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
			return
		case errors.Is(err, models.ErrConversationForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		default:
			observability.Error(r.Context(), "chat.conversations.get.failed", map[string]any{
				"user_id":         userID,
				"conversation_id": conversationID,
				"error":           err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load conversation"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"conversation": conversation})
}

func (h ConversationHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	conversationID, valid := validateConversationID(w, r)
	if !valid {
		return
	}

	page, limit, err := parsePaginationParams(r, 1, maxConversationDetailMessagePage)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	res, err := h.Conversations.ListMessages(r.Context(), conversationID, userID, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrConversationNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
			return
		case errors.Is(err, models.ErrConversationForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		default:
			observability.Error(r.Context(), "chat.messages.list.failed", map[string]any{
				"user_id":         userID,
				"conversation_id": conversationID,
				"error":           err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load messages"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages":    res.Messages,
		"total":       res.Total,
		"page":        res.Page,
		"limit":       res.Limit,
		"total_pages": res.TotalPages,
	})
}

func (h ConversationHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	conversationID, valid := validateConversationID(w, r)
	if !valid {
		return
	}

	var req sendMessageRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if claimedSenderID := strings.TrimSpace(req.SenderID); claimedSenderID != "" && claimedSenderID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "text is required"})
		return
	}
	if utf8.RuneCountInString(text) > maxConversationMessageLen {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "text is too long"})
		return
	}

	message, delivery, err := h.Conversations.AddMessage(r.Context(), conversationID, userID, text)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrConversationNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
			return
		case errors.Is(err, models.ErrConversationForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		default:
			observability.Error(r.Context(), "chat.messages.create.failed", map[string]any{
				"user_id":         userID,
				"conversation_id": conversationID,
				"error":           err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send message"})
			return
		}
	}

	h.sendNewMessageEmail(r, delivery)
	writeJSON(w, http.StatusCreated, map[string]any{"message": message})
}

func (h ConversationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	conversationID, valid := validateConversationID(w, r)
	if !valid {
		return
	}

	if err := h.Conversations.MarkRead(r.Context(), conversationID, userID); err != nil {
		switch {
		case errors.Is(err, models.ErrConversationNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
			return
		case errors.Is(err, models.ErrConversationForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		default:
			observability.Error(r.Context(), "chat.messages.mark_read.failed", map[string]any{
				"user_id":         userID,
				"conversation_id": conversationID,
				"error":           err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark conversation read"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "marked read"})
}

func validateConversationID(w http.ResponseWriter, r *http.Request) (string, bool) {
	conversationID := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(conversationID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid conversation id"})
		return "", false
	}
	return conversationID, true
}

func parsePaginationParams(r *http.Request, defaultPage, defaultLimit int) (int, int, error) {
	page := defaultPage
	limit := defaultLimit

	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return 0, 0, errors.New("page must be a positive integer")
		}
		page = n
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return 0, 0, errors.New("limit must be a positive integer")
		}
		limit = n
	}

	return page, limit, nil
}

func (h ConversationHandler) sendNewMessageEmail(r *http.Request, delivery *models.MessageNotificationDelivery) {
	if h.EmailSender == nil || delivery == nil || strings.TrimSpace(delivery.RecipientEmail) == "" {
		return
	}

	subject := delivery.Notification.Title
	body := delivery.Notification.Body
	if strings.TrimSpace(body) == "" {
		body = "You have a new message on ReSellution."
	}
	if err := h.EmailSender.Send(delivery.RecipientEmail, subject, body); err != nil {
		observability.Warn(r.Context(), "notifications.email.new_message.failed", map[string]any{
			"notification_id": delivery.Notification.ID,
			"user_id":         delivery.Notification.UserID,
			"error":           err.Error(),
		})
	}
}
