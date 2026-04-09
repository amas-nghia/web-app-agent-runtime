package tarotweb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSeedDeckHas78Cards(t *testing.T) {
	t.Parallel()
	if got := len(seedDeck()); got != 78 {
		t.Fatalf("len(seedDeck()) = %d, want 78", got)
	}
}

func TestHomeAndReading(t *testing.T) {
	t.Parallel()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Moonlight Tarot") {
		t.Fatalf("GET / body missing title: %s", rec.Body.String())
	}

	form := strings.NewReader("question=What+should+I+focus+on%3F")
	req = httptest.NewRequest(http.MethodPost, "/read", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /read status = %d, want 303", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/reading/") {
		t.Fatalf("POST /read location = %q, want /reading/...", loc)
	}
}
