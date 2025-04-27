package api

import (
	"context"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hionay/quotes/internal/domain"
)

type mockRepo struct {
	GetLatestFunc    func(ctx context.Context, page, limit int) ([]*domain.Quote, error)
	GetTopFunc       func(ctx context.Context, page, limit int) ([]*domain.Quote, error)
	GetRandomFunc    func(ctx context.Context) (*domain.Quote, error)
	CreateFunc       func(ctx context.Context, q *domain.Quote) error
	LikeQuoteFunc    func(ctx context.Context, id int) error
	DislikeQuoteFunc func(ctx context.Context, id int) error
	GetByIDFunc      func(ctx context.Context, id int) (*domain.Quote, error)
}

func (m *mockRepo) GetLatest(ctx context.Context, page, limit int) ([]*domain.Quote, error) {
	return m.GetLatestFunc(ctx, page, limit)
}
func (m *mockRepo) GetTop(ctx context.Context, page, limit int) ([]*domain.Quote, error) {
	return m.GetTopFunc(ctx, page, limit)
}
func (m *mockRepo) GetRandom(ctx context.Context) (*domain.Quote, error) {
	return m.GetRandomFunc(ctx)
}
func (m *mockRepo) Create(ctx context.Context, q *domain.Quote) error {
	return m.CreateFunc(ctx, q)
}
func (m *mockRepo) LikeQuote(ctx context.Context, id int) error {
	return m.LikeQuoteFunc(ctx, id)
}
func (m *mockRepo) DislikeQuote(ctx context.Context, id int) error {
	return m.DislikeQuoteFunc(ctx, id)
}
func (m *mockRepo) GetByID(ctx context.Context, id int) (*domain.Quote, error) {
	return m.GetByIDFunc(ctx, id)
}

func TestParsePage(t *testing.T) {
	tests := []struct {
		query  string
		expect int
	}{
		{"", 1},
		{"?page=abc", 1},
		{"?page=0", 1},
		{"?page=3", 3},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
		p := parsePage(r)
		if p != tt.expect {
			t.Errorf("parsePage(%q) = %d; want %d", tt.query, p, tt.expect)
		}
	}
}

func TestParseVote(t *testing.T) {
	tests := []struct {
		url     string
		wantID  int
		wantTyp string
		errOK   bool
	}{
		{"/vote", 0, "", false},
		{"/vote?id=1", 0, "", false},
		{"/vote?type=up", 0, "", false},
		{"/vote?id=x&type=up", 0, "", false},
		{"/vote?id=5&type=up", 5, "up", true},
		{"/vote?id=8&type=down", 8, "down", true},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, tt.url, nil)
		id, vt, err := parseVote(r)
		if tt.errOK {
			if err != nil {
				t.Errorf("parseVote(%q) unexpected error: %v", tt.url, err)
				continue
			}
			if id != tt.wantID || vt != tt.wantTyp {
				t.Errorf("parseVote(%q) = (%d,%q); want (%d,%q)", tt.url, id, vt, tt.wantID, tt.wantTyp)
			}
		} else {
			if err == nil {
				t.Errorf("parseVote(%q) expected error, got nil", tt.url)
			}
		}
	}
}

func TestListHandler(t *testing.T) {
	repo := &mockRepo{
		GetLatestFunc: func(ctx context.Context, page, limit int) ([]*domain.Quote, error) {
			return []*domain.Quote{{ID: 1, Quote: "q1"}, {ID: 2, Quote: "q2"}}, nil
		},
	}
	a := &API{
		logger:    slog.Default(),
		quoteRepo: repo,
		tmpl:      template.Must(template.New("index.html").Parse(`{{len .Quotes}} quotes`)),
	}
	r := httptest.NewRequest(http.MethodGet, "/?page=2", nil)
	w := httptest.NewRecorder()
	h := a.listHandler(repo.GetLatest)
	h(w, r)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want %d", res.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "2 quotes") {
		t.Errorf("body = %q; want contains %q", string(body), "2 quotes")
	}
}

func TestAddQuote(t *testing.T) {
	var created *domain.Quote
	repo := &mockRepo{
		CreateFunc: func(ctx context.Context, q *domain.Quote) error {
			created = q
			return nil
		},
	}
	a := &API{quoteRepo: repo, logger: slog.Default()}
	r := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader("quote=hello+world&comment=nice"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	a.addQuote(w, r)
	res := w.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d; want %d", res.StatusCode, http.StatusSeeOther)
	}
	if created == nil {
		t.Fatal("expected quote to be created, got nil")
	}
	if created.Quote != "hello world" {
		t.Errorf("created.Quote = %q; want %q", created.Quote, "hello world")
	}
	if created.Comment != "nice" {
		t.Errorf("created.Comment = %q; want %q", created.Comment, "nice")
	}
}

func TestVoteHandler(t *testing.T) {
	called := false
	repo := &mockRepo{
		LikeQuoteFunc: func(ctx context.Context, id int) error {
			if id != 7 {
				return errors.New("unexpected id")
			}
			called = true
			return nil
		},
		GetByIDFunc: func(ctx context.Context, id int) (*domain.Quote, error) {
			return &domain.Quote{ID: id, Quote: "q"}, nil
		},
	}
	a := &API{
		logger:    slog.Default(),
		quoteRepo: repo,
		tmpl:      template.Must(template.New("quote-card.html").Parse(`quote {{.ID}}`)),
	}

	r := httptest.NewRequest(http.MethodGet, "/vote?id=7&type=up", nil)
	w := httptest.NewRecorder()
	a.voteHandler(w, r)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want %d", res.StatusCode, http.StatusOK)
	}
	if !called {
		t.Error("expected LikeQuote to be called")
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "quote 7") {
		t.Errorf("body = %q; want contains %q", string(body), "quote 7")
	}
}

func TestViewHandler(t *testing.T) {
	repo := &mockRepo{
		GetByIDFunc: func(ctx context.Context, id int) (*domain.Quote, error) {
			if id != 42 {
				return nil, errors.New("not found")
			}
			return &domain.Quote{ID: 42, Quote: "x"}, nil
		},
	}
	a := &API{
		logger:    slog.Default(),
		quoteRepo: repo,
		tmpl:      template.Must(template.New("index.html").Parse(`ID={{(index .Quotes 0).ID}}`)),
	}

	r1 := httptest.NewRequest(http.MethodGet, "/quote/42", nil)
	w1 := httptest.NewRecorder()
	a.viewHandler(w1, r1)
	rsp1 := w1.Result()
	if rsp1.StatusCode != http.StatusOK {
		t.Errorf("status = %d; want %d", rsp1.StatusCode, http.StatusOK)
	}
	b1, _ := io.ReadAll(w1.Body)
	if !strings.Contains(string(b1), "ID=42") {
		t.Errorf("body = %q; want contains %q", string(b1), "ID=42")
	}

	r2 := httptest.NewRequest(http.MethodGet, "/quote/abc", nil)
	w2 := httptest.NewRecorder()
	a.viewHandler(w2, r2)
	rsp2 := w2.Result()
	if rsp2.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rsp2.StatusCode, http.StatusBadRequest)
	}
}
