package tarotweb

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type App struct {
	lib   *Library
	tmpls *template.Template
	mux   *http.ServeMux
}

func New() (*App, error) {
	tmpls, err := parseTemplates()
	if err != nil {
		return nil, err
	}

	app := &App{
		lib:   NewLibrary(),
		tmpls: tmpls,
		mux:   http.NewServeMux(),
	}
	app.routes()
	return app, nil
}

func (a *App) Handler() http.Handler { return a.mux }

func (a *App) routes() {
	a.mux.HandleFunc("GET /", a.home)
	a.mux.HandleFunc("POST /read", a.createReading)
	a.mux.HandleFunc("GET /reading/", a.viewReading)
	a.mux.HandleFunc("GET /cards", a.cards)
	a.mux.HandleFunc("GET /api/cards", a.apiCards)
	a.mux.HandleFunc("GET /api/readings/", a.apiReading)
	a.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))
}

type homeView struct {
	Title    string
	Cards    []Card
	Reading  *Reading
	Message  string
	Question string
	Year     int
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.notFound(w, r)
		return
	}
	data := homeView{
		Title:   "Moonlight Tarot",
		Cards:   a.lib.Cards()[:12],
		Year:    time.Now().Year(),
		Message: "Draw a three-card spread and explore the full deck.",
	}
	a.render(w, "home", data)
}

func (a *App) createReading(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.renderError(w, http.StatusBadRequest, "Unable to read your question")
		return
	}
	question := strings.TrimSpace(r.FormValue("question"))
	reading := a.lib.Reading(question)
	http.Redirect(w, r, "/reading/"+reading.ID, http.StatusSeeOther)
}

func (a *App) viewReading(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/reading/")
	reading, ok := a.lib.ReadingByID(id)
	if !ok {
		a.renderError(w, http.StatusNotFound, "Reading not found")
		return
	}
	a.render(w, "reading", reading)
}

func (a *App) cards(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cards" {
		a.notFound(w, r)
		return
	}
	data := struct {
		Title string
		Cards []Card
	}{Title: "The Tarot Deck", Cards: a.lib.Cards()}
	a.render(w, "cards", data)
}

func (a *App) apiCards(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.lib.Cards())
}

func (a *App) apiReading(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/readings/")
	reading, ok := a.lib.ReadingByID(id)
	if !ok {
		a.renderError(w, http.StatusNotFound, "Reading not found")
		return
	}
	writeJSON(w, reading)
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tmpls.ExecuteTemplate(w, name, data); err != nil {
		a.renderError(w, http.StatusInternalServerError, fmt.Sprintf("render %s: %v", name, err))
	}
}

func (a *App) renderError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = a.tmpls.ExecuteTemplate(w, "error", struct {
		Status  int
		Message string
	}{Status: status, Message: message})
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	a.renderError(w, http.StatusNotFound, "Page not found")
}
