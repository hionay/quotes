package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hionay/quotes/internal/domain"
)

const (
	quoteFields = "id, quote, comment, date, ip, likes, votes"
	baseSelect  = "SELECT " + quoteFields + " FROM quotes"
)

type QuoteRepository struct {
	db Connection
}

func NewQuoteRepository(db Connection) *QuoteRepository {
	return &QuoteRepository{db: db}
}

func (qr *QuoteRepository) Create(ctx context.Context, q *domain.Quote) error {
	const insertQuery = `
        INSERT INTO quotes (quote, comment, date, ip)
        VALUES (?, ?, ?, ?)
    `
	if _, err := qr.db.ExecContext(ctx,
		insertQuery,
		q.Quote, q.Comment, q.Date, q.IP,
	); err != nil {
		return fmt.Errorf("insert quote: %w", err)
	}
	return nil
}

func (qr *QuoteRepository) GetByID(ctx context.Context, id int) (*domain.Quote, error) {
	query := baseSelect + " WHERE id = ?"
	row := qr.db.QueryRowContext(ctx, query, id)
	return scanQuote(row)
}

func (qr *QuoteRepository) GetLatest(ctx context.Context, page, limit int) ([]*domain.Quote, error) {
	query := baseSelect + " ORDER BY date DESC LIMIT ? OFFSET ?"
	return qr.queryQuotes(ctx, query, limit, (page-1)*limit)
}

func (qr *QuoteRepository) GetTop(ctx context.Context, page, limit int) ([]*domain.Quote, error) {
	query := baseSelect + " ORDER BY likes DESC LIMIT ? OFFSET ?"
	return qr.queryQuotes(ctx, query, limit, (page-1)*limit)
}

func (qr *QuoteRepository) GetRandom(ctx context.Context) (*domain.Quote, error) {
	query := baseSelect + " ORDER BY RAND() LIMIT 1"
	row := qr.db.QueryRowContext(ctx, query)
	return scanQuote(row)
}

func (qr *QuoteRepository) LikeQuote(ctx context.Context, id int) error {
	const updateQuery = `
		UPDATE quotes
		SET likes = likes + 1, votes = votes + 1
		WHERE id = ?
	`
	if _, err := qr.db.ExecContext(ctx, updateQuery, id); err != nil {
		return fmt.Errorf("update quote: %w", err)
	}
	return nil
}

func (qr *QuoteRepository) DislikeQuote(ctx context.Context, id int) error {
	const updateQuery = `
		UPDATE quotes
		SET likes = likes - 1, votes = votes + 1
		WHERE id = ?
	`
	if _, err := qr.db.ExecContext(ctx, updateQuery, id); err != nil {
		return fmt.Errorf("update quote: %w", err)
	}
	return nil
}

func (qr *QuoteRepository) queryQuotes(ctx context.Context, query string, args ...any) ([]*domain.Quote, error) {
	rows, err := qr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query quotes: %w", err)
	}
	defer rows.Close()

	var list []*domain.Quote
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return list, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanQuote(s scanner) (*domain.Quote, error) {
	var q domain.Quote
	var rawDate string
	if err := s.Scan(
		&q.ID, &q.Quote, &q.Comment, &rawDate,
		&q.IP, &q.Likes, &q.Votes,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quote not found: %w", err)
		}
		return nil, fmt.Errorf("scan quote: %w", err)
	}
	q.Date = parseMySQLDate(rawDate)
	return &q, nil
}
