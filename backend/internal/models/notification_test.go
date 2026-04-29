package models

import "testing"

func TestNormalizeNotificationPagination(t *testing.T) {
	page, limit := normalizeNotificationPagination(0, 0)
	if page != 1 || limit != defaultNotificationListLimit {
		t.Fatalf("expected page=1 limit=%d, got page=%d limit=%d", defaultNotificationListLimit, page, limit)
	}

	page, limit = normalizeNotificationPagination(5, 999)
	if page != 5 || limit != maxNotificationListLimit {
		t.Fatalf("expected page=5 limit=%d, got page=%d limit=%d", maxNotificationListLimit, page, limit)
	}
}

func TestCalculateNotificationTotalPages(t *testing.T) {
	if got := calculateNotificationTotalPages(0, 20); got != 1 {
		t.Fatalf("expected 1 page for no records, got %d", got)
	}
	if got := calculateNotificationTotalPages(101, 20); got != 6 {
		t.Fatalf("expected 6 pages, got %d", got)
	}
}

