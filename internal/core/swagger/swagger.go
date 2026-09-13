// Package swagger mounts the Swagger UI.
//
// The UI is only registered for non-production environments; in production
// Register is a no-op so the docs and UI are unreachable. Generated spec
// files live in docs (see make swag) and must be regenerated after
// annotation changes.
package swagger

import (
	_ "embed"
	"net/http"
	"strings"

	_ "github.com/fiqrikm18/quran-app/internal/core/swagger/docs"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//go:embed SwaggerDark.css
var swaggerDarkCSS []byte

// swaggerDarkFile is served under /swagger/ alongside the UI.
const swaggerDarkFile = "SwaggerDark.css"

// darkThemeInjector appends a <link> for the vendored SwaggerDark stylesheet.
// It runs via http-swagger's BeforeScript hook (inside window.onload, before
// the UI object is created), so the theme applies on first paint without any
// external network dependency. Load failures are reported to the browser
// console so a missing/broken stylesheet can't fail silently.
const darkThemeInjector = `(function(){var l=document.createElement('link');l.rel='stylesheet';l.type='text/css';l.href='./SwaggerDark.css';l.onload=function(){console.info('[swagger] SwaggerDark theme loaded');};l.onerror=function(){console.error('[swagger] failed to load SwaggerDark theme from '+l.href);};document.head.appendChild(l);})();`

// Register mounts /swagger/* on r unless env is production.
// Serve/write failures are logged to the terminal via logger.
// It returns true when the UI was mounted.
func Register(r *chi.Mux, env string, logger zerolog.Logger) bool {
	if strings.ToLower(strings.TrimSpace(env)) == "production" {
		return false
	}

	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		// Keep the Authorize token across page refreshes.
		httpSwagger.PersistAuthorization(true),
		// SwaggerDark theme (vendored, always-dark variant).
		httpSwagger.BeforeScript(darkThemeInjector),
		// NOTE: do not pass raw URLs via UIConfig — http-swagger renders
		// UIConfig values verbatim into the page JS, so a URL without
		// embedded quotes is a syntax error that blanks the whole UI.
	)

	r.Get("/swagger/*", func(w http.ResponseWriter, req *http.Request) {
		// Serve the vendored SwaggerDark theme; http-swagger's file server
		// only knows its own bundled assets, so intercept this one path.
		if strings.HasSuffix(req.URL.Path, "/"+swaggerDarkFile) {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			if _, err := w.Write(swaggerDarkCSS); err != nil {
				logger.Error().
					Err(err).
					Str("file", swaggerDarkFile).
					Str("method", req.Method).
					Str("path", req.URL.Path).
					Str("remote", req.RemoteAddr).
					Msg("failed to serve swagger dark theme")
			}
			return
		}
		swaggerHandler.ServeHTTP(w, req)
	})
	return true
}
