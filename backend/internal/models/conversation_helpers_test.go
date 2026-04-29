package models

import (
	"strings"
	"testing"
)

func TestNormalizeConversationPagination(t *testing.T) {
	page, limit := normalizeConversationPagination(0, 0)
	if page != 1 || limit != defaultConversationListLimit {
		t.Fatalf("expected page=1 limit=%d, got page=%d limit=%d", defaultConversationListLimit, page, limit)
	}

	page, limit = normalizeConversationPagination(2, 999)
	if page != 2 || limit != maxConversationListLimit {
		t.Fatalf("expected page=2 limit=%d, got page=%d limit=%d", maxConversationListLimit, page, limit)
	}
}

func TestNormalizeMessagePagination(t *testing.T) {
	page, limit := normalizeMessagePagination(-3, -1)
	if page != 1 || limit != defaultConversationMessageLimit {
		t.Fatalf("expected page=1 limit=%d, got page=%d limit=%d", defaultConversationMessageLimit, page, limit)
	}

	page, limit = normalizeMessagePagination(3, 500)
	if page != 3 || limit != maxConversationMessageLimit {
		t.Fatalf("expected page=3 limit=%d, got page=%d limit=%d", maxConversationMessageLimit, page, limit)
	}
}

func TestCalculateTotalPages(t *testing.T) {
	if got := calculateTotalPages(0, 10); got != 1 {
		t.Fatalf("expected 1 for empty total, got %d", got)
	}
	if got := calculateTotalPages(21, 10); got != 3 {
		t.Fatalf("expected 3 pages, got %d", got)
	}
}

func TestBuildMessageReadBy(t *testing.T) {
	got := buildMessageReadBy("buyer", "buyer", "seller", false)
	if len(got) != 1 || got[0] != "buyer" {
		t.Fatalf("expected only sender when unread, got %#v", got)
	}

	got = buildMessageReadBy("buyer", "buyer", "seller", true)
	if len(got) != 2 || got[0] != "buyer" || got[1] != "seller" {
		t.Fatalf("expected buyer and seller read list, got %#v", got)
	}
}

func TestMessageNotificationParticipants(t *testing.T) {
	summary := conversationSummary{
		Conversation: Conversation{
			BuyerID:    "buyer-id",
			SellerID:   "seller-id",
			SellerName: "Seller Name",
		},
		buyerName: "Buyer Name",
	}

	recipientID, senderName, recipientName := messageNotificationParticipants(summary, "buyer-id")
	if recipientID != "seller-id" || senderName != "Buyer Name" || recipientName != "Seller Name" {
		t.Fatalf("unexpected buyer->seller mapping: %s %s %s", recipientID, senderName, recipientName)
	}

	recipientID, senderName, recipientName = messageNotificationParticipants(summary, "unknown")
	if recipientID != "" || senderName != "" || recipientName != "" {
		t.Fatalf("expected empty values for unknown sender, got: %s %s %s", recipientID, senderName, recipientName)
	}
}

func TestBuildMessageNotificationBody(t *testing.T) {
	body := buildMessageNotificationBody("Alex", "Desk", "  ")
	if !strings.Contains(body, "Alex sent you a new message about Desk.") {
		t.Fatalf("unexpected base body: %q", body)
	}

	longText := strings.Repeat("x", 250)
	body = buildMessageNotificationBody("Alex", "Desk", longText)
	if len(body) == 0 || !strings.Contains(body, "...") {
		t.Fatalf("expected truncated snippet with ellipsis, got: %q", body)
	}
}

