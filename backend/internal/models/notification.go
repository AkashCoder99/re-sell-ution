package models

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"time"
)

var ErrNotificationNotFound = errors.New("notification not found")

const (
	defaultNotificationListLimit = 20
	maxNotificationListLimit     = 50
)

type Notification struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Body        string    `json:"body,omitempty"`
	ReferenceID *string   `json:"reference_id,omitempty"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type NotificationsPage struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	Limit         int            `json:"limit"`
	TotalPages    int            `json:"total_pages"`
	UnreadTotal   int            `json:"unread_total"`
}

type NotificationStore struct {
	DB *sql.DB
}

func (s NotificationStore) ListByUser(ctx context.Context, userID string, unreadOnly *bool, page, limit int) (NotificationsPage, error) {
	page, limit = normalizeNotificationPagination(page, limit)

	args := []any{userID}
	where := "user_id = $1"
	if unreadOnly != nil {
		args = append(args, *unreadOnly)
		where += " AND is_read = $2"
	}

	var total int
	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE `+where, args...).Scan(&total); err != nil {
		return NotificationsPage{}, err
	}

	var unreadTotal int
	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1
		  AND is_read = FALSE
	`, userID).Scan(&unreadTotal); err != nil {
		return NotificationsPage{}, err
	}

	offset := (page - 1) * limit
	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, user_id, type, title, body, reference_id, is_read, created_at
		FROM notifications
		WHERE `+where+`
		ORDER BY created_at DESC, id ASC
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return NotificationsPage{}, err
	}
	defer rows.Close()

	notifications := make([]Notification, 0, limit)
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return NotificationsPage{}, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return NotificationsPage{}, err
	}

	return NotificationsPage{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    calculateNotificationTotalPages(total, limit),
		UnreadTotal:   unreadTotal,
	}, nil
}

func (s NotificationStore) MarkRead(ctx context.Context, notificationID, userID string) (Notification, error) {
	row := s.DB.QueryRowContext(ctx, `
		UPDATE notifications
		SET is_read = TRUE
		WHERE id = $1
		  AND user_id = $2
		RETURNING id, user_id, type, title, body, reference_id, is_read, created_at
	`, notificationID, userID)

	notification, err := scanNotification(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Notification{}, ErrNotificationNotFound
		}
		return Notification{}, err
	}

	return notification, nil
}

func (s NotificationStore) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = TRUE
		WHERE user_id = $1
		  AND is_read = FALSE
	`, userID)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

type notificationScanner interface {
	Scan(dest ...any) error
}

func scanNotification(scanner notificationScanner) (Notification, error) {
	var notification Notification
	var body sql.NullString
	var referenceID sql.NullString

	err := scanner.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&body,
		&referenceID,
		&notification.IsRead,
		&notification.CreatedAt,
	)
	if err != nil {
		return Notification{}, err
	}

	if body.Valid {
		notification.Body = body.String
	}
	if referenceID.Valid {
		v := referenceID.String
		notification.ReferenceID = &v
	}

	return notification, nil
}

func normalizeNotificationPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultNotificationListLimit
	}
	if limit > maxNotificationListLimit {
		limit = maxNotificationListLimit
	}
	return page, limit
}

func calculateNotificationTotalPages(total, limit int) int {
	if total <= 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
