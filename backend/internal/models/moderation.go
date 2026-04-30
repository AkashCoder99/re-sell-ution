package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ReportTargetListing = "listing"
	ReportTargetUser    = "user"
	ReportTargetMessage = "message"

	ReportStatusOpen     = "open"
	ReportStatusInReview = "in_review"
	ReportStatusResolved = "resolved"
	ReportStatusRejected = "rejected"
)

var ErrReportNotFound = errors.New("report not found")
var ErrInvalidReportTransition = errors.New("invalid report status transition")
var ErrListingReportOwnListing = errors.New("cannot report your own listing")

type Report struct {
	ID                string             `json:"id"`
	ReporterUserID    string             `json:"reporter_user_id"`
	TargetType        string             `json:"target_type"`
	TargetListingID   *string            `json:"target_listing_id,omitempty"`
	TargetUserID      *string            `json:"target_user_id,omitempty"`
	TargetMessageID   *string            `json:"target_message_id,omitempty"`
	ReasonCode        string             `json:"reason_code,omitempty"`
	ReasonText        string             `json:"reason_text,omitempty"`
	Status            string             `json:"status"`
	Priority          int                `json:"priority"`
	AssignedAdminID   *string            `json:"assigned_admin_id,omitempty"`
	ResolutionNote    string             `json:"resolution_note,omitempty"`
	ResolvedAt        *time.Time         `json:"resolved_at,omitempty"`
	RetainUntil       *time.Time         `json:"retain_until,omitempty"`
	PurgeAfter        *time.Time         `json:"purge_after,omitempty"`
	IsLegalHold       bool               `json:"is_legal_hold"`
	LegalHoldReason   string             `json:"legal_hold_reason,omitempty"`
	DeletedAt         *time.Time         `json:"deleted_at,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	ModerationActions []ModerationAction `json:"moderation_actions,omitempty"`
}

type ModerationAction struct {
	ID            string          `json:"id"`
	ReportID      string          `json:"report_id"`
	ActorUserID   string          `json:"actor_user_id"`
	ActionType    string          `json:"action_type"`
	ActionPayload json.RawMessage `json:"action_payload,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type CreateReportInput struct {
	ReporterUserID  string
	TargetType      string
	TargetListingID *string
	TargetUserID    *string
	TargetMessageID *string
	ReasonCode      string
	ReasonText      string
	Priority        int
}

type ListReportsFilters struct {
	Status          string
	AssignedAdminID string
	TargetType      string
	ReporterUserID  string
}

type ReportsPage struct {
	Reports    []Report `json:"reports"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}

type ReportStore struct {
	DB *sql.DB
}

func (s ReportStore) Create(ctx context.Context, in CreateReportInput) (Report, error) {
	if in.Priority < 1 || in.Priority > 5 {
		in.Priority = 3
	}
	in.ReasonCode = strings.TrimSpace(in.ReasonCode)
	in.ReasonText = strings.TrimSpace(in.ReasonText)

	var report Report
	report.ID = uuid.NewString()
	report.ReporterUserID = in.ReporterUserID
	report.TargetType = strings.TrimSpace(in.TargetType)
	report.TargetListingID = in.TargetListingID
	report.TargetUserID = in.TargetUserID
	report.TargetMessageID = in.TargetMessageID
	report.ReasonCode = in.ReasonCode
	report.ReasonText = in.ReasonText
	report.Status = ReportStatusOpen
	report.Priority = in.Priority

	query := `
		INSERT INTO reports (
			id, reporter_user_id, target_type, target_listing_id, target_user_id, target_message_id,
			reason_code, reason_text, status, priority
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING created_at, updated_at
	`
	if err := s.DB.QueryRowContext(
		ctx,
		query,
		report.ID,
		report.ReporterUserID,
		report.TargetType,
		report.TargetListingID,
		report.TargetUserID,
		report.TargetMessageID,
		report.ReasonCode,
		report.ReasonText,
		report.Status,
		report.Priority,
	).Scan(&report.CreatedAt, &report.UpdatedAt); err != nil {
		return Report{}, err
	}
	return report, nil
}

func (s ReportStore) CreateListingReport(ctx context.Context, listingID, reporterID, reason string) (Report, error) {
	var sellerID string
	err := s.DB.QueryRowContext(ctx, `
		SELECT seller_id
		FROM listings
		WHERE id = $1
		  AND deleted_at IS NULL
	`, listingID).Scan(&sellerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, ErrListingNotFound
		}
		return Report{}, err
	}
	if sellerID == reporterID {
		return Report{}, ErrListingReportOwnListing
	}

	listingIDCopy := listingID
	report, err := s.Create(ctx, CreateReportInput{
		ReporterUserID:  reporterID,
		TargetType:      ReportTargetListing,
		TargetListingID: &listingIDCopy,
		ReasonText:      reason,
		Priority:        3,
	})
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate") || strings.Contains(lowerErr, "unique") {
			return s.findExistingListingReport(ctx, listingID, reporterID)
		}
		return Report{}, err
	}
	return report, nil
}

func (s ReportStore) findExistingListingReport(ctx context.Context, listingID, reporterID string) (Report, error) {
	var reportID string
	err := s.DB.QueryRowContext(ctx, `
		SELECT id
		FROM reports
		WHERE reporter_user_id = $1
		  AND target_type = 'listing'
		  AND target_listing_id = $2
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, reporterID, listingID).Scan(&reportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, ErrReportNotFound
		}
		return Report{}, err
	}
	return s.GetReport(ctx, reportID)
}

func (s ReportStore) List(ctx context.Context, filters ListReportsFilters, page, limit int) (ReportsPage, error) {
	page, limit = normalizeReportPagination(page, limit)
	whereSQL, args := buildReportListWhere(filters)

	var total int
	countQuery := "SELECT COUNT(*) FROM reports r " + whereSQL
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return ReportsPage{}, err
	}

	offset := (page - 1) * limit
	args = append(args, limit, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)

	listQuery := fmt.Sprintf(`
		SELECT
			id, reporter_user_id, target_type, target_listing_id, target_user_id, target_message_id,
			reason_code, reason_text, status, priority, assigned_admin_id, resolution_note, resolved_at,
			retain_until, purge_after, is_legal_hold, legal_hold_reason, deleted_at, created_at, updated_at
		FROM reports r
		%s
		ORDER BY created_at DESC, id ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, limitPos, offsetPos)

	rows, err := s.DB.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return ReportsPage{}, err
	}
	defer rows.Close()

	reports := make([]Report, 0, limit)
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return ReportsPage{}, err
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return ReportsPage{}, err
	}

	return ReportsPage{
		Reports:    reports,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calculateReportTotalPages(total, limit),
	}, nil
}

func (s ReportStore) ListReports(ctx context.Context, status string, page, limit int) (ReportsPage, error) {
	if strings.TrimSpace(status) == "all" {
		status = ""
	}
	return s.List(ctx, ListReportsFilters{Status: status}, page, limit)
}

func (s ReportStore) GetReport(ctx context.Context, reportID string) (Report, error) {
	var report Report
	err := s.DB.QueryRowContext(ctx, `
		SELECT
			id, reporter_user_id, target_type, target_listing_id, target_user_id, target_message_id,
			reason_code, reason_text, status, priority, assigned_admin_id, resolution_note, resolved_at,
			retain_until, purge_after, is_legal_hold, legal_hold_reason, deleted_at, created_at, updated_at
		FROM reports
		WHERE id = $1
		  AND deleted_at IS NULL
	`, reportID).Scan(
		&report.ID,
		&report.ReporterUserID,
		&report.TargetType,
		&report.TargetListingID,
		&report.TargetUserID,
		&report.TargetMessageID,
		&report.ReasonCode,
		&report.ReasonText,
		&report.Status,
		&report.Priority,
		&report.AssignedAdminID,
		&report.ResolutionNote,
		&report.ResolvedAt,
		&report.RetainUntil,
		&report.PurgeAfter,
		&report.IsLegalHold,
		&report.LegalHoldReason,
		&report.DeletedAt,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, ErrReportNotFound
		}
		return Report{}, err
	}
	actions, err := s.ListActions(ctx, reportID)
	if err != nil {
		return Report{}, err
	}
	report.ModerationActions = actions
	return report, nil
}

func (s ReportStore) UpdateReportStatus(ctx context.Context, reportID, adminID, status, resolutionNote string) (Report, error) {
	status = strings.TrimSpace(status)
	switch status {
	case ReportStatusInReview:
		return s.Assign(ctx, reportID, adminID, adminID)
	case ReportStatusOpen:
		return s.transitionWithAction(ctx, reportID, adminID, status, resolutionNote, "", "note", map[string]any{
			"status":          status,
			"resolution_note": strings.TrimSpace(resolutionNote),
		})
	case ReportStatusResolved, ReportStatusRejected:
		return s.Resolve(ctx, reportID, adminID, status, resolutionNote)
	default:
		return Report{}, ErrInvalidReportTransition
	}
}

func (s ReportStore) Assign(ctx context.Context, reportID, adminID, actorUserID string) (Report, error) {
	return s.transitionWithAction(ctx, reportID, actorUserID, ReportStatusInReview, "", adminID, "assign", map[string]any{
		"assigned_admin_id": adminID,
	})
}

func (s ReportStore) Resolve(ctx context.Context, reportID, actorUserID, newStatus, resolutionNote string) (Report, error) {
	if strings.TrimSpace(newStatus) == "" {
		newStatus = ReportStatusResolved
	}
	actionType := "close_report"
	return s.transitionWithAction(ctx, reportID, actorUserID, newStatus, resolutionNote, "", actionType, map[string]any{
		"status":          newStatus,
		"resolution_note": strings.TrimSpace(resolutionNote),
	})
}

func (s ReportStore) AddAction(ctx context.Context, reportID, actorUserID, actionType string, payload map[string]any) (ModerationAction, error) {
	action := ModerationAction{
		ID:          uuid.NewString(),
		ReportID:    reportID,
		ActorUserID: actorUserID,
		ActionType:  strings.TrimSpace(actionType),
	}
	if len(payload) == 0 {
		action.ActionPayload = json.RawMessage(`{}`)
	} else {
		raw, err := json.Marshal(payload)
		if err != nil {
			return ModerationAction{}, err
		}
		action.ActionPayload = raw
	}
	if err := s.DB.QueryRowContext(ctx, `
		INSERT INTO moderation_actions (id, report_id, actor_user_id, action_type, action_payload)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING created_at
	`, action.ID, action.ReportID, action.ActorUserID, action.ActionType, action.ActionPayload).Scan(&action.CreatedAt); err != nil {
		return ModerationAction{}, err
	}
	return action, nil
}

func (s ReportStore) ListActions(ctx context.Context, reportID string) ([]ModerationAction, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, report_id, actor_user_id, action_type, action_payload, created_at
		FROM moderation_actions
		WHERE report_id = $1
		ORDER BY created_at DESC, id ASC
	`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]ModerationAction, 0)
	for rows.Next() {
		var action ModerationAction
		if err := rows.Scan(
			&action.ID,
			&action.ReportID,
			&action.ActorUserID,
			&action.ActionType,
			&action.ActionPayload,
			&action.CreatedAt,
		); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, rows.Err()
}

func (s ReportStore) transitionWithAction(ctx context.Context, reportID, actorUserID, newStatus, resolutionNote, assignedAdminID, actionType string, payload map[string]any) (Report, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback()

	var currentStatus string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM reports WHERE id = $1`, reportID).Scan(&currentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, ErrReportNotFound
		}
		return Report{}, err
	}
	if err := validateReportStatusTransition(currentStatus, newStatus); err != nil {
		return Report{}, err
	}

	var report Report
	err = tx.QueryRowContext(ctx, `
		UPDATE reports
		SET
			status = $2,
			assigned_admin_id = CASE WHEN $3 = '' THEN assigned_admin_id ELSE $3::uuid END,
			resolution_note = CASE WHEN $4 = '' THEN resolution_note ELSE $4 END,
			resolved_at = CASE WHEN $2 IN ('resolved', 'rejected') THEN NOW() ELSE resolved_at END,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, reporter_user_id, target_type, target_listing_id, target_user_id, target_message_id,
			reason_code, reason_text, status, priority, assigned_admin_id, resolution_note, resolved_at,
			retain_until, purge_after, is_legal_hold, legal_hold_reason, deleted_at, created_at, updated_at
	`, reportID, newStatus, strings.TrimSpace(assignedAdminID), strings.TrimSpace(resolutionNote)).Scan(
		&report.ID,
		&report.ReporterUserID,
		&report.TargetType,
		&report.TargetListingID,
		&report.TargetUserID,
		&report.TargetMessageID,
		&report.ReasonCode,
		&report.ReasonText,
		&report.Status,
		&report.Priority,
		&report.AssignedAdminID,
		&report.ResolutionNote,
		&report.ResolvedAt,
		&report.RetainUntil,
		&report.PurgeAfter,
		&report.IsLegalHold,
		&report.LegalHoldReason,
		&report.DeletedAt,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, ErrReportNotFound
		}
		return Report{}, err
	}

	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return Report{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO moderation_actions (id, report_id, actor_user_id, action_type, action_payload)
		VALUES ($1,$2,$3,$4,$5)
	`, uuid.NewString(), reportID, actorUserID, actionType, payloadRaw); err != nil {
		return Report{}, err
	}

	if err := tx.Commit(); err != nil {
		return Report{}, err
	}
	return report, nil
}

func validateReportStatusTransition(current, next string) error {
	current = strings.TrimSpace(current)
	next = strings.TrimSpace(next)
	if current == "" || next == "" {
		return ErrInvalidReportTransition
	}
	if current == next {
		return nil
	}
	allowed := map[string]map[string]struct{}{
		ReportStatusOpen: {
			ReportStatusInReview: {},
			ReportStatusResolved: {},
			ReportStatusRejected: {},
		},
		ReportStatusInReview: {
			ReportStatusResolved: {},
			ReportStatusRejected: {},
			ReportStatusOpen:     {},
		},
		ReportStatusResolved: {},
		ReportStatusRejected: {},
	}
	if _, ok := allowed[current][next]; !ok {
		return ErrInvalidReportTransition
	}
	return nil
}

func buildReportListWhere(filters ListReportsFilters) (string, []any) {
	clauses := []string{"r.deleted_at IS NULL"}
	args := make([]any, 0, 4)
	if v := strings.TrimSpace(filters.Status); v != "" {
		args = append(args, v)
		clauses = append(clauses, fmt.Sprintf("r.status = $%d", len(args)))
	}
	if v := strings.TrimSpace(filters.AssignedAdminID); v != "" {
		args = append(args, v)
		clauses = append(clauses, fmt.Sprintf("r.assigned_admin_id = $%d", len(args)))
	}
	if v := strings.TrimSpace(filters.TargetType); v != "" {
		args = append(args, v)
		clauses = append(clauses, fmt.Sprintf("r.target_type = $%d", len(args)))
	}
	if v := strings.TrimSpace(filters.ReporterUserID); v != "" {
		args = append(args, v)
		clauses = append(clauses, fmt.Sprintf("r.reporter_user_id = $%d", len(args)))
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

type reportScanner interface {
	Scan(dest ...any) error
}

func scanReport(scanner reportScanner) (Report, error) {
	var r Report
	var reasonCode, reasonText, resolutionNote, legalHoldReason sql.NullString
	err := scanner.Scan(
		&r.ID,
		&r.ReporterUserID,
		&r.TargetType,
		&r.TargetListingID,
		&r.TargetUserID,
		&r.TargetMessageID,
		&reasonCode,
		&reasonText,
		&r.Status,
		&r.Priority,
		&r.AssignedAdminID,
		&resolutionNote,
		&r.ResolvedAt,
		&r.RetainUntil,
		&r.PurgeAfter,
		&r.IsLegalHold,
		&legalHoldReason,
		&r.DeletedAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err != nil {
		return Report{}, err
	}
	if reasonCode.Valid {
		r.ReasonCode = reasonCode.String
	}
	if reasonText.Valid {
		r.ReasonText = reasonText.String
	}
	if resolutionNote.Valid {
		r.ResolutionNote = resolutionNote.String
	}
	if legalHoldReason.Valid {
		r.LegalHoldReason = legalHoldReason.String
	}
	return r, nil
}

func normalizeReportPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func calculateReportTotalPages(total, limit int) int {
	if total <= 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
