package api

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hionay/quotes/internal/config"
	"github.com/hionay/quotes/internal/domain"
	"github.com/hionay/quotes/internal/repository"
)

var tagRe = regexp.MustCompile(`<([^<]+)>`)

type API struct {
	srv       *http.Server
	logger    *slog.Logger
	quoteRepo domain.QuoteRepository
	tmpl      *template.Template
}

func NewAPI(cfg *config.Config, logger *slog.Logger, db repository.Connection) *API {
	tmpl := template.Must(template.New("").ParseGlob("templates/*.html"))
	api := &API{
		logger:    logger,
		quoteRepo: repository.NewQuoteRepository(db),
		tmpl:      tmpl,
	}

	mux := http.NewServeMux()
	mux.Handle("/", api.listHandler(api.quoteRepo.GetLatest))
	mux.Handle("/top", api.listHandler(api.quoteRepo.GetTop))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/random", api.randomHandler)
	mux.HandleFunc("/add", api.addQuote)
	mux.HandleFunc("/vote", api.voteHandler)
	mux.HandleFunc("/quote/", api.viewHandler)

	api.srv = &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.ServerPort()),
		Handler:      mux,
		ReadTimeout:  apiReadTimeout,
		WriteTimeout: apiWriteTimeout,
		IdleTimeout:  apiIdleTimeout,
	}
	return api
}

func (a *API) ListenAndServe() error {
	err := a.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	a.logger.Info("Server stopped")
	return nil
}

func (a *API) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), apiShutdownTimeout)
	defer cancel()
	return a.srv.Shutdown(ctx)
}

func (a *API) listHandler(
	fetch func(ctx context.Context, page, limit int) ([]*domain.Quote, error),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := parsePage(r)
		q, err := fetch(r.Context(), page, defaultLimit)
		if err != nil {
			a.error(w, http.StatusInternalServerError, "fetching quotes", err)
			return
		}
		vms := toViewModels(q)
		a.render(w, "index.html", map[string]any{
			"Quotes":   vms,
			"HasPrev":  page > 1,
			"HasNext":  len(q) == defaultLimit,
			"PrevPage": page - 1,
			"NextPage": page + 1,
			"Endpoint": path.Clean(r.URL.Path),
		})
	}
}

func (a *API) randomHandler(w http.ResponseWriter, r *http.Request) {
	quote, err := a.quoteRepo.GetRandom(r.Context())
	if err != nil {
		a.error(w, http.StatusInternalServerError, "fetching random quote", err)
		return
	}
	vms := toViewModels([]*domain.Quote{quote})
	a.render(w, "index.html", map[string]any{"Quotes": vms})
}

func (a *API) voteHandler(w http.ResponseWriter, r *http.Request) {
	id, vote, err := parseVote(r)
	if err != nil {
		a.error(w, http.StatusBadRequest, "invalid vote request", err)
		return
	}
	action := map[string]func(context.Context, int) error{
		"up":   a.quoteRepo.LikeQuote,
		"down": a.quoteRepo.DislikeQuote,
	}[vote]

	if err := action(r.Context(), id); err != nil {
		a.error(w, http.StatusInternalServerError, "applying vote", err)
		return
	}

	quote, err := a.quoteRepo.GetByID(r.Context(), id)
	if err != nil {
		a.error(w, http.StatusInternalServerError, "fetching updated quote", err)
		return
	}
	vm := toViewModels([]*domain.Quote{quote})[0]
	a.render(w, "quote-card.html", vm)
}

func (a *API) viewHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.TrimPrefix(r.URL.Path, "/quote/")
	id, err := strconv.Atoi(parts)
	if err != nil {
		a.error(w, http.StatusBadRequest, "invalid quote ID", err)
		return
	}
	quote, err := a.quoteRepo.GetByID(r.Context(), id)
	if err != nil {
		a.error(w, http.StatusNotFound, "quote not found", err)
		return
	}
	vms := toViewModels([]*domain.Quote{quote})
	a.render(w, "index.html", map[string]any{"Quotes": vms})
}

func (a *API) addQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.error(w, http.StatusMethodNotAllowed, "use POST to add quote", nil)
		return
	}
	quote := &domain.Quote{
		Quote:   nl2br(r.FormValue("quote")),
		Comment: nl2br(r.FormValue("comment")),
		Date:    time.Now(),
		IP:      r.RemoteAddr,
	}
	if err := a.quoteRepo.Create(r.Context(), quote); err != nil {
		a.error(w, http.StatusInternalServerError, "adding quote", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *API) render(w http.ResponseWriter, tpl string, data any) {
	if err := a.tmpl.ExecuteTemplate(w, tpl, data); err != nil {
		a.error(w, http.StatusInternalServerError, "rendering "+tpl, err)
	}
}

func (a *API) error(w http.ResponseWriter, status int, msg string, err error) {
	http.Error(w, msg, status)
	if err != nil {
		a.logger.Error(msg, slog.Any("error", err))
	}
}

func parsePage(r *http.Request) int {
	if p := r.URL.Query().Get("page"); p != "" {
		if i, err := strconv.Atoi(p); err == nil && i > 0 {
			return i
		}
	}
	return 1
}

func parseVote(r *http.Request) (id int, vote string, _ error) {
	q := r.URL.Query()
	vote = q.Get("type")
	if vote != "up" && vote != "down" {
		return 0, "", errors.New("must be up or down")
	}
	idStr := q.Get("id")
	id, err := strconv.Atoi(idStr)
	return id, vote, err
}

func toViewModels(quotes []*domain.Quote) []Quote {
	vms := make([]Quote, len(quotes))
	for i, q := range quotes {
		vms[i] = Quote{
			ID:      q.ID,
			Quote:   sanitize(q.Quote),
			Comment: sanitize(q.Comment),
			Date:    q.Date,
			IP:      q.IP,
			Likes:   q.Likes,
			Votes:   q.Votes,
		}
	}
	return vms
}

func nl2br(s string) string {
	return strings.ReplaceAll(s, "\n", "<br />")
}

func sanitize(raw string) template.HTML {
	r := strings.NewReplacer(
		"<br />", "<br>",
		`<br \/>`, "<br>",
	)
	raw = r.Replace(raw)
	return template.HTML(tagRe.ReplaceAllStringFunc(raw, func(m string) string {
		if m == "<br>" {
			return m
		}
		return template.HTMLEscapeString(m)
	}))
}
