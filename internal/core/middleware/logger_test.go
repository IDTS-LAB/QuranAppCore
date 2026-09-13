package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func serveThroughChain(logger zerolog.Logger, req *http.Request) *httptest.ResponseRecorder {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	rec := httptest.NewRecorder()
	chimw.RequestID(StructuredLogger(logger)(next)).ServeHTTP(rec, req)
	return rec
}

func TestStructuredLoggerLogsRequestCompletion(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	serveThroughChain(logger, req)

	out := buf.String()
	if !strings.Contains(out, `"status":418`) {
		t.Fatalf("expected request completion log with status 418, got: %q", out)
	}
	if !strings.Contains(out, `"path":"/api/healthz"`) {
		t.Fatalf("expected log to contain request path, got: %q", out)
	}
}

func TestRequestIDGeneratedAndEchoed(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := serveThroughChain(logger, req)

	out := buf.String()
	reqID := rec.Header().Get(chimw.RequestIDHeader)
	if reqID == "" {
		t.Fatalf("expected %s response header to be set", chimw.RequestIDHeader)
	}
	if !strings.Contains(out, `"req_id":"`+reqID+`"`) {
		t.Fatalf("expected log req_id %q to match response header, got: %q", reqID, out)
	}
}

func TestIncomingRequestIDHonored(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	req.Header.Set(chimw.RequestIDHeader, "client-provided-id")
	rec := serveThroughChain(logger, req)

	if got := rec.Header().Get(chimw.RequestIDHeader); got != "client-provided-id" {
		t.Fatalf("expected incoming request ID to be echoed, got: %q", got)
	}
	if out := buf.String(); !strings.Contains(out, `"req_id":"client-provided-id"`) {
		t.Fatalf("expected log req_id to honor incoming ID, got: %q", out)
	}
}
