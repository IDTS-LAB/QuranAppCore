package swagger

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func newTestRouter(env string) *chi.Mux {
	return newTestRouterWithLogger(env, zerolog.New(io.Discard))
}

func newTestRouterWithLogger(env string, logger zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()
	Register(r, env, logger)
	return r
}

func TestRegisterDisabledInProduction(t *testing.T) {
	r := chi.NewRouter()
	if Register(r, "production", zerolog.New(io.Discard)) {
		t.Fatal("expected Register to be a no-op in production")
	}
}

func TestSwaggerIndexInjectsDarkTheme(t *testing.T) {
	r := newTestRouter("development")

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for index.html, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "SwaggerDark.css") {
		t.Fatal("expected index.html to reference SwaggerDark.css injector")
	}
}

// failWriter simulates a broken connection so w.Write returns an error.
type failWriter struct {
	header http.Header
}

func (f *failWriter) Header() http.Header { return f.header }
func (f *failWriter) WriteHeader(int)     {}
func (f *failWriter) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}

func TestSwaggerDarkCSSWriteErrorLogsWithZerolog(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	r := newTestRouterWithLogger("development", logger)

	req := httptest.NewRequest(http.MethodGet, "/swagger/SwaggerDark.css", nil)
	r.ServeHTTP(&failWriter{header: http.Header{}}, req)

	out := buf.String()
	if !strings.Contains(out, "failed to serve swagger dark theme") {
		t.Fatalf("expected zerolog error for CSS write failure, got: %q", out)
	}
	if !strings.Contains(out, "boom") {
		t.Fatalf("expected logged error to contain write cause, got: %q", out)
	}
}
func TestSwaggerDarkCSSServed(t *testing.T) {
	r := newTestRouter("development")

	req := httptest.NewRequest(http.MethodGet, "/swagger/SwaggerDark.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for SwaggerDark.css, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Fatalf("expected text/css content type, got %q", ct)
	}
	if !strings.Contains(rec.Body.String(), ".swagger-ui") {
		t.Fatal("expected SwaggerDark.css body to contain swagger-ui rules")
	}
}
