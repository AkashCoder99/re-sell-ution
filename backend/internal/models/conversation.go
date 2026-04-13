package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrConversationNotFound = errors.New("conversation not found")
var ErrConversationForbidden = errors.New("conversation access forbidden")
var ErrConversationSelfNotAllowed = errors.New("cannot create a conversation on your own listing")

const (
	defaultConversationListLimit    = 8
	maxConversationListLimit        = 50
	defaultConversationMessageLimit = 50
	maxConversationMessageLimit     = 100
)

type Message struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	ReadBy    []string  `json:"read_by,omitempty"`
}

type Conversation struct {
	ID              string    `json:"id"`
	ListingID       string    `json:"listing_id"`
	ListingTitle    string    `json:"listing_title"`
	ListingPrice    float64   `json:"listing_price"`
	ListingCity     string    `json:"listing_city"`
	BuyerID         string    `json:"buyer_id"`
	SellerID        string    `json:"seller_id"`
	SellerName      string    `json:"seller_name"`
	ParticipantName string    `json:"participant_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	UnreadCount     int       `json:"unread_count"`
	LastMessageText string    `json:"last_message_text"`
	Messages        []Message `json:"messages"`
}

type ConversationsPage struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	Limit         int            `json:"limit"`
	TotalPages    int            `json:"total_pages"`
}

type MessagesPage struct {
	Messages   []Message `json:"messages"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type ConversationStore struct {
	DB *sql.DB
}

type conversationSummary struct {
	Conversation
	buyerName string
}

func (s ConversationStore) CreateOrGet(ctx context.Context, listingID, buyerID string) (Conversation, error) {
	var sellerID string
	err := s.DB.QueryRowContext(ctx, `
		SELECT seller_id
		FROM listings
		WHERE id = $1
		  AND deleted_at IS NULL
	`, listingID).Scan(&sellerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, ErrListingNotFound
		}
		return Conversation{}, err
	}

	if sellerID == buyerID {
		return Conversation{}, ErrConversationSelfNotAllowed
	}

	conversationID := uuid.NewString()
	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO conversations (id, listing_id, buyer_id, seller_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (listing_id, buyer_id) DO NOTHING
	`, conversationID, listingID, buyerID, sellerID); err != nil {
		return Conversation{}, err
	}

	err = s.DB.QueryRowContext(ctx, `
		SELECT id
		FROM conversations
		WHERE listing_id = $1
		  AND buyer_id = $2
	`, listingID, buyerID).Scan(&conversationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, err
	}

	return s.GetByID(ctx, conversationID, buyerID, 1, defaultConversationMessageLimit)
}

func (s ConversationStore) ListByUser(ctx context.Context, userID, query string, page, limit int) (ConversationsPage, error) {
	page, limit = normalizeConversationPagination(page, limit)
	query = strings.TrimSpace(query)

	args := []any{userID}
	where := []string{"(c.buyer_id = $1 OR c.seller_id = $1)"}
	if query != "" {
		args = append(args, "%"+query+"%")
		searchPos := len(args)
		where = append(where, fmt.Sprintf(`(
			l.title ILIKE $%[1]d OR
			seller.full_name ILIKE $%[1]d OR
			buyer.full_name ILIKE $%[1]d OR
			EXISTS (
				SELECT 1
				FROM messages m
				WHERE m.conversation_id = c.id
				  AND m.body ILIKE $%[1]d
			)
		)`, searchPos))
	}

	fromClause := `
		FROM conversations c
		JOIN listings l ON l.id = c.listing_id
		JOIN users seller ON seller.id = c.seller_id
		JOIN users buyer ON buyer.id = c.buyer_id
	`
	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) %s WHERE %s`, fromClause, whereClause)
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return ConversationsPage{}, err
	}

	offset := (page - 1) * limit
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	listArgs := append(append([]any{}, args...), limit, offset)

	querySQL := fmt.Sprintf(`
		SELECT
			c.id,
			c.listing_id,
			l.title,
			l.price,
			l.city,
			c.buyer_id,
			c.seller_id,
			seller.full_name,
			buyer.full_name,
			c.created_at,
			c.updated_at,
			COALESCE(last_message.body, '') AS last_message_text,
			COALESCE(unread.unread_count, 0) AS unread_count
		%s
		LEFT JOIN LATERAL (
			SELECT m.body
			FROM messages m
			WHERE m.conversation_id = c.id
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT 1
		) last_message ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS unread_count
			FROM messages m
			WHERE m.conversation_id = c.id
			  AND m.sender_id <> $1
			  AND m.is_read = FALSE
		) unread ON true
		WHERE %s
		ORDER BY COALESCE(c.last_message_at, c.updated_at, c.created_at) DESC, c.id ASC
		LIMIT $%d OFFSET $%d
	`, fromClause, whereClause, limitPos, offsetPos)

	rows, err := s.DB.QueryContext(ctx, querySQL, listArgs...)
	if err != nil {
		return ConversationsPage{}, err
	}
	defer rows.Close()

	conversations := make([]Conversation, 0, limit)
	for rows.Next() {
		summary, err := scanConversationSummary(rows, userID)
		if err != nil {
			return ConversationsPage{}, err
		}
		conversations = append(conversations, summary.Conversation)
	}
	if err := rows.Err(); err != nil {
		return ConversationsPage{}, err
	}

	return ConversationsPage{
		Conversations: conversations,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    calculateTotalPages(total, limit),
	}, nil
}

func (s ConversationStore) GetByID(ctx context.Context, conversationID, userID string, page, limit int) (Conversation, error) {
	summary, err := s.loadConversationSummary(ctx, conversationID, userID)
	if err != nil {
		return Conversation{}, err
	}

	messagePage, err := s.listMessagesPage(ctx, summary, page, limit)
	if err != nil {
		return Conversation{}, err
	}

	summary.Conversation.Messages = messagePage.Messages
	return summary.Conversation, nil
}

func (s ConversationStore) ListMessages(ctx context.Context, conversationID, userID string, page, limit int) (MessagesPage, error) {
	summary, err := s.loadConversationSummary(ctx, conversationID, userID)
	if err != nil {
		return MessagesPage{}, err
	}

	return s.listMessagesPage(ctx, summary, page, limit)
}

func (s ConversationStore) AddMessage(ctx context.Context, conversationID, userID, text string) (Message, error) {
	summary, err := s.loadConversationSummary(ctx, conversationID, userID)
	if err != nil {
		return Message{}, err
	}
	if userID != summary.BuyerID && userID != summary.SellerID {
		return Message{}, ErrConversationForbidden
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, err
	}
	defer tx.Rollback()

	messageID := uuid.NewString()
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `
		INSERT INTO messages (id, conversation_id, sender_id, body, is_read)
		VALUES ($1, $2, $3, $4, FALSE)
		RETURNING created_at
	`, messageID, conversationID, userID, text).Scan(&createdAt)
	if err != nil {
		return Message{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE conversations
		SET updated_at = $2, last_message_at = $2
		WHERE id = $1
	`, conversationID, createdAt); err != nil {
		return Message{}, err
	}

	if err := tx.Commit(); err != nil {
		return Message{}, err
	}

	return Message{
		ID:        messageID,
		SenderID:  userID,
		Text:      text,
		CreatedAt: createdAt,
		ReadBy:    []string{userID},
	}, nil
}

func (s ConversationStore) MarkRead(ctx context.Context, conversationID, userID string) error {
	summary, err := s.loadConversationSummary(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if userID != summary.BuyerID && userID != summary.SellerID {
		return ErrConversationForbidden
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE messages
		SET is_read = TRUE
		WHERE conversation_id = $1
		  AND sender_id <> $2
		  AND is_read = FALSE
	`, conversationID, userID)
	return err
}

func (s ConversationStore) loadConversationSummary(ctx context.Context, conversationID, userID string) (conversationSummary, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT
			c.id,
			c.listing_id,
			l.title,
			l.price,
			l.city,
			c.buyer_id,
			c.seller_id,
			seller.full_name,
			buyer.full_name,
			c.created_at,
			c.updated_at,
			COALESCE(last_message.body, '') AS last_message_text,
			COALESCE(unread.unread_count, 0) AS unread_count
		FROM conversations c
		JOIN listings l ON l.id = c.listing_id
		JOIN users seller ON seller.id = c.seller_id
		JOIN users buyer ON buyer.id = c.buyer_id
		LEFT JOIN LATERAL (
			SELECT m.body
			FROM messages m
			WHERE m.conversation_id = c.id
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT 1
		) last_message ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS unread_count
			FROM messages m
			WHERE m.conversation_id = c.id
			  AND m.sender_id <> $2
			  AND m.is_read = FALSE
		) unread ON true
		WHERE c.id = $1
	`, conversationID, userID)

	summary, err := scanConversationSummary(row, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return conversationSummary{}, ErrConversationNotFound
		}
		return conversationSummary{}, err
	}

	if userID != summary.BuyerID && userID != summary.SellerID {
		return conversationSummary{}, ErrConversationForbidden
	}

	return summary, nil
}

func (s ConversationStore) listMessagesPage(ctx context.Context, summary conversationSummary, page, limit int) (MessagesPage, error) {
	page, limit = normalizeMessagePagination(page, limit)

	var total int
	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM messages
		WHERE conversation_id = $1
	`, summary.ID).Scan(&total); err != nil {
		return MessagesPage{}, err
	}

	offset := (page - 1) * limit
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, sender_id, body, is_read, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, summary.ID, limit, offset)
	if err != nil {
		return MessagesPage{}, err
	}
	defer rows.Close()

	messagesDesc := make([]Message, 0, limit)
	for rows.Next() {
		var message Message
		var isRead bool
		if err := rows.Scan(&message.ID, &message.SenderID, &message.Text, &isRead, &message.CreatedAt); err != nil {
			return MessagesPage{}, err
		}
		message.ReadBy = buildMessageReadBy(message.SenderID, summary.BuyerID, summary.SellerID, isRead)
		messagesDesc = append(messagesDesc, message)
	}
	if err := rows.Err(); err != nil {
		return MessagesPage{}, err
	}

	messages := reverseMessages(messagesDesc)
	return MessagesPage{
		Messages:   messages,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calculateTotalPages(total, limit),
	}, nil
}

type conversationScanner interface {
	Scan(dest ...any) error
}

func scanConversationSummary(scanner conversationScanner, userID string) (conversationSummary, error) {
	var summary conversationSummary
	var sellerName string
	err := scanner.Scan(
		&summary.ID,
		&summary.ListingID,
		&summary.ListingTitle,
		&summary.ListingPrice,
		&summary.ListingCity,
		&summary.BuyerID,
		&summary.SellerID,
		&sellerName,
		&summary.buyerName,
		&summary.CreatedAt,
		&summary.UpdatedAt,
		&summary.LastMessageText,
		&summary.UnreadCount,
	)
	if err != nil {
		return conversationSummary{}, err
	}

	summary.SellerName = sellerName
	summary.ParticipantName = sellerName
	if userID == summary.SellerID {
		summary.ParticipantName = summary.buyerName
	}
	summary.Messages = make([]Message, 0)
	return summary, nil
}

func normalizeConversationPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultConversationListLimit
	}
	if limit > maxConversationListLimit {
		limit = maxConversationListLimit
	}
	return page, limit
}

func normalizeMessagePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultConversationMessageLimit
	}
	if limit > maxConversationMessageLimit {
		limit = maxConversationMessageLimit
	}
	return page, limit
}

func calculateTotalPages(total, limit int) int {
	if total <= 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}

func reverseMessages(messages []Message) []Message {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages
}

func buildMessageReadBy(senderID, buyerID, sellerID string, isRead bool) []string {
	readBy := []string{senderID}
	if !isRead {
		return readBy
	}

	switch senderID {
	case buyerID:
		if sellerID != "" && sellerID != senderID {
			readBy = append(readBy, sellerID)
		}
	case sellerID:
		if buyerID != "" && buyerID != senderID {
			readBy = append(readBy, buyerID)
		}
	}

	return readBy
}
