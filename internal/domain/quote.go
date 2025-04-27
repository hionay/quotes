package domain

import (
	"context"
	"time"
)

type Quote struct {
	Date    time.Time
	Quote   string
	Comment string
	IP      string
	ID      int
	Likes   int
	Votes   int
}

type QuoteRepository interface {
	Create(context.Context, *Quote) error
	GetByID(context.Context, int) (*Quote, error)
	GetLatest(context.Context, int, int) ([]*Quote, error)
	GetTop(context.Context, int, int) ([]*Quote, error)
	GetRandom(context.Context) (*Quote, error)
	LikeQuote(context.Context, int) error
	DislikeQuote(context.Context, int) error
}
