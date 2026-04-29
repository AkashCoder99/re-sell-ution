package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

type stubConversationStore struct {
	createOrGetFn  func(ctx context.Context, listingID, buyerID string) (models.Conversation, error)
	listByUserFn   func(ctx context.Context, userID, query string, page, limit int) (models.ConversationsPage, error)
	getByIDFn      func(ctx context.Context, conversationID, userID string, page, limit int) (models.Conversation, error)
	listMessagesFn func(ctx context.Context, conversationID, userID string, page, limit int) (models.MessagesPage, error)
	addMessageFn   func(ctx context.Context, conversationID, userID, text string) (models.Message, *models.MessageNotificationDelivery, error)
	markReadFn     func(ctx context.Context, conversationID, userID string) error
}

func (s stubConversationStore) CreateOrGet(ctx context.Context, listingID, buyerID string) (models.Conversation, error) {
	if s.createOrGetFn == nil {
		return models.Conversation{}, nil
	}
	return s.createOrGetFn(ctx, listingID, buyerID)
}

func (s stubConversationStore) ListByUser(ctx context.Context, userID, query string, page, limit int) (models.ConversationsPage, error) {
	if s.listByUserFn == nil {
		return models.ConversationsPage{}, nil
	}
	return s.listByUserFn(ctx, userID, query, page, limit)
}

func (s stubConversationStore) GetByID(ctx context.Context, conversationID, userID string, page, limit int) (models.Conversation, error) {
	if s.getByIDFn == nil {
		return models.Conversation{}, nil
	}
	return s.getByIDFn(ctx, conversationID, userID, page, limit)
}

func (s stubConversationStore) ListMessages(ctx context.Context, conversationID, userID string, page, limit int) (models.MessagesPage, error) {
	if s.listMessagesFn == nil {
		return models.MessagesPage{}, nil
	}
	return s.listMessagesFn(ctx, conversationID, userID, page, limit)
}

func (s stubConversationStore) AddMessage(ctx context.Context, conversationID, userID, text string) (models.Message, *models.MessageNotificationDelivery, error) {
	if s.addMessageFn == nil {
		return models.Message{}, nil, nil
	}
	return s.addMessageFn(ctx, conversationID, userID, text)
}

func (s stubConversationStore) MarkRead(ctx context.Context, conversationID, userID string) error {
	if s.markReadFn == nil {
		return nil
	}
	return s.markReadFn(ctx, conversationID, userID)
}

func TestConversationHandlerListUnauthorized(t *testing.T) {
	h := ConversationHandler{}
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/chat/conversations", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestConversationHandlerCreateUnauthorized(t *testing.T) {
	h := ConversationHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations", strings.NewReader(`{}`))
	h.Create(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestConversationHandlerCreateInvalidListingID(t *testing.T) {
	const secret = "conversation-http-test-secret-1"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations", strings.NewReader(`{"listing_id":"bad-id"}`))
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "12121212-1212-1212-1212-121212121212"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerCreateBuyerMismatchForbidden(t *testing.T) {
	const secret = "conversation-http-test-secret-2"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations", strings.NewReader(`{
		"listing_id":"123e4567-e89b-12d3-a456-426614174000",
		"buyer_id":"00000000-0000-0000-0000-000000000999"
	}`))
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "13131313-1313-1313-1313-131313131313"))
	wrapped(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestConversationHandlerListInvalidPagination(t *testing.T) {
	const secret = "conversation-http-test-secret-3"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/conversations?page=abc", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "14141414-1414-1414-1414-141414141414"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerGetInvalidConversationID(t *testing.T) {
	const secret = "conversation-http-test-secret-4"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/conversations/bad-id", nil)
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "15151515-1515-1515-1515-151515151515"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerGetForbidden(t *testing.T) {
	const secret = "conversation-http-test-secret-5"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{
		Conversations: stubConversationStore{
			getByIDFn: func(ctx context.Context, conversationID, userID string, page, limit int) (models.Conversation, error) {
				return models.Conversation{}, models.ErrConversationForbidden
			},
		},
	}
	wrapped := middleware.Auth(tm, h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000", nil)
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "16161616-1616-1616-1616-161616161616"))
	wrapped(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestConversationHandlerListMessagesInvalidPagination(t *testing.T) {
	const secret = "conversation-http-test-secret-6"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.ListMessages)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000/messages?limit=oops", nil)
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "17171717-1717-1717-1717-171717171717"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerSendMessageInvalidJSON(t *testing.T) {
	const secret = "conversation-http-test-secret-7"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.SendMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000/messages", strings.NewReader(`{`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "18181818-1818-1818-1818-181818181818"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerSendMessageEmptyText(t *testing.T) {
	const secret = "conversation-http-test-secret-8"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.SendMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000/messages", strings.NewReader(`{"text":"   "}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "19191919-1919-1919-1919-191919191919"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerSendMessageSenderMismatchForbidden(t *testing.T) {
	const secret = "conversation-http-test-secret-9"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.SendMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000/messages", strings.NewReader(`{
		"sender_id":"00000000-0000-0000-0000-000000000001",
		"text":"hello"
	}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "20202020-2020-2020-2020-202020202020"))
	wrapped(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestConversationHandlerSendMessageNotFound(t *testing.T) {
	const secret = "conversation-http-test-secret-10"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{
		Conversations: stubConversationStore{
			addMessageFn: func(ctx context.Context, conversationID, userID, text string) (models.Message, *models.MessageNotificationDelivery, error) {
				return models.Message{}, nil, models.ErrConversationNotFound
			},
		},
	}
	wrapped := middleware.Auth(tm, h.SendMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000/messages", strings.NewReader(`{"text":"hello"}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "21212121-2121-2121-2121-212121212121"))
	wrapped(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestConversationHandlerMarkReadInvalidConversationID(t *testing.T) {
	const secret = "conversation-http-test-secret-11"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{}
	wrapped := middleware.Auth(tm, h.MarkRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/conversations/bad-id/read", nil)
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "22222222-2222-2222-2222-222222222222"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConversationHandlerMarkReadForbidden(t *testing.T) {
	const secret = "conversation-http-test-secret-12"
	tm := utils.NewTokenManager(secret)
	h := ConversationHandler{
		Conversations: stubConversationStore{
			markReadFn: func(ctx context.Context, conversationID, userID string) error {
				return models.ErrConversationForbidden
			},
		},
	}
	wrapped := middleware.Auth(tm, h.MarkRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/conversations/123e4567-e89b-12d3-a456-426614174000/read", nil)
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "23232323-2323-2323-2323-232323232323"))
	wrapped(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
