package models

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateReportStatusTransition(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		wantErr bool
	}{
		{name: "open to in_review", current: ReportStatusOpen, next: ReportStatusInReview, wantErr: false},
		{name: "open to resolved", current: ReportStatusOpen, next: ReportStatusResolved, wantErr: false},
		{name: "in_review to open", current: ReportStatusInReview, next: ReportStatusOpen, wantErr: false},
		{name: "resolved to open disallowed", current: ReportStatusResolved, next: ReportStatusOpen, wantErr: true},
		{name: "rejected to in_review disallowed", current: ReportStatusRejected, next: ReportStatusInReview, wantErr: true},
		{name: "same status allowed", current: ReportStatusOpen, next: ReportStatusOpen, wantErr: false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateReportStatusTransition(tc.current, tc.next)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantErr && !errors.Is(err, ErrInvalidReportTransition) {
				t.Fatalf("expected ErrInvalidReportTransition, got %v", err)
			}
		})
	}
}

func TestBuildReportListWhere(t *testing.T) {
	where, args := buildReportListWhere(ListReportsFilters{
		Status:          "open",
		AssignedAdminID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		TargetType:      ReportTargetListing,
		ReporterUserID:  "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	})
	if !strings.Contains(where, "r.status = $1") {
		t.Fatalf("missing status clause: %s", where)
	}
	if !strings.Contains(where, "r.assigned_admin_id = $2") {
		t.Fatalf("missing assigned clause: %s", where)
	}
	if !strings.Contains(where, "r.target_type = $3") {
		t.Fatalf("missing target type clause: %s", where)
	}
	if !strings.Contains(where, "r.reporter_user_id = $4") {
		t.Fatalf("missing reporter clause: %s", where)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(args))
	}
}

func TestNormalizeReportPagination(t *testing.T) {
	page, limit := normalizeReportPagination(0, 0)
	if page != 1 || limit != 20 {
		t.Fatalf("expected page=1 limit=20, got page=%d limit=%d", page, limit)
	}
	page, limit = normalizeReportPagination(2, 999)
	if page != 2 || limit != 100 {
		t.Fatalf("expected page=2 limit=100, got page=%d limit=%d", page, limit)
	}
}

func TestCalculateReportTotalPages(t *testing.T) {
	if got := calculateReportTotalPages(0, 20); got != 1 {
		t.Fatalf("expected 1 got %d", got)
	}
	if got := calculateReportTotalPages(51, 20); got != 3 {
		t.Fatalf("expected 3 got %d", got)
	}
}

