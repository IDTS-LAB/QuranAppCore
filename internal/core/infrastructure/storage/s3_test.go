package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/config"
	"github.com/rs/zerolog"
)

// Live MinIO round-trip. Runs against the dev MinIO from
// docker-compose.dev.yml (make minio-up); skips when it isn't reachable so
// plain `go test ./...` stays green without infra.
func testStore(t *testing.T) *S3 {
	t.Helper()

	conn, err := net.DialTimeout("tcp", "127.0.0.1:9000", 2*time.Second)
	if err != nil {
		t.Skip("MinIO not reachable at 127.0.0.1:9000 (run: make minio-up)")
	}
	_ = conn.Close()

	logger := zerolog.New(io.Discard)
	store, err := New(config.S3Config{
		Endpoint:     "http://127.0.0.1:9000",
		Region:       "us-east-1",
		AccessKey:    "minioadmin",
		SecretKey:    "minioadmin",
		Bucket:       "quran-test",
		UsePathStyle: true,
		PresignTTL:   "15m",
	}, logger)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return store
}

func TestS3RoundTripAgainstMinIO(t *testing.T) {
	store := testStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := fmt.Sprintf("roundtrip/%d.txt", time.Now().UnixNano())
	payload := []byte("hello minio")

	if err := store.Put(ctx, key, payload, "text/plain"); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	obj, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	got, err := io.ReadAll(obj.Body)
	_ = obj.Body.Close()
	if err != nil {
		t.Fatalf("reading object body: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("expected %q, got %q", payload, got)
	}
	if obj.ContentType != "text/plain" {
		t.Fatalf("expected content-type text/plain, got %q", obj.ContentType)
	}

	url, err := store.PresignedGetURL(ctx, key, 0)
	if err != nil {
		t.Fatalf("PresignedGetURL returned error: %v", err)
	}
	if !strings.Contains(url, "X-Amz-Signature=") {
		t.Fatalf("expected signed URL, got %q", url)
	}
	resp, err := http.Get(url) //nolint:gosec,noctx // test-only, short-lived URL
	if err != nil {
		t.Fatalf("GET presigned URL: %v", err)
	}
	signed, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("reading presigned body: %v", err)
	}
	if resp.StatusCode != http.StatusOK || string(signed) != string(payload) {
		t.Fatalf("presigned download failed: status=%d body=%q", resp.StatusCode, signed)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := store.Get(ctx, key); err == nil {
		t.Fatal("expected Get after Delete to fail, got nil")
	}
}

func TestS3BatchPresignAgainstMinIO(t *testing.T) {
	store := testStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prefix := fmt.Sprintf("batch/%d", time.Now().UnixNano())
	payloads := map[string]string{
		prefix + "/a.txt": "alpha",
		prefix + "/b.txt": "beta",
		prefix + "/c.txt": "gamma",
	}
	keys := make([]string, 0, len(payloads))
	for key, body := range payloads {
		if err := store.Put(ctx, key, []byte(body), "text/plain"); err != nil {
			t.Fatalf("Put %q: %v", key, err)
		}
		keys = append(keys, key)
	}

	urls, err := store.PresignedGetURLs(ctx, keys, 0)
	if err != nil {
		t.Fatalf("PresignedGetURLs returned error: %v", err)
	}
	if len(urls) != len(keys) {
		t.Fatalf("expected %d URLs, got %d", len(keys), len(urls))
	}
	for _, key := range keys {
		url, ok := urls[key]
		if !ok || !strings.Contains(url, "X-Amz-Signature=") {
			t.Fatalf("missing or unsigned URL for %q: %q", key, url)
		}
		resp, err := http.Get(url) //nolint:gosec,noctx // test-only, short-lived URL
		if err != nil {
			t.Fatalf("GET presigned URL for %q: %v", key, err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("reading presigned body for %q: %v", key, err)
		}
		if resp.StatusCode != http.StatusOK || string(body) != payloads[key] {
			t.Fatalf("presigned download mismatch for %q: status=%d body=%q", key, resp.StatusCode, body)
		}
		if err := store.Delete(ctx, key); err != nil {
			t.Fatalf("Delete %q: %v", key, err)
		}
	}
}

func TestS3PresignedUploadAgainstMinIO(t *testing.T) {
	store := testStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	put := func(url, contentType string, body []byte) {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
		if err != nil {
			t.Fatalf("building PUT request: %v", err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		// MinIO/S3 answers presigned PUTs with 200.
		req.ContentLength = int64(len(body))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PUT presigned URL: %v", err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("presigned upload failed: status=%d url=%q", resp.StatusCode, url)
		}
	}
	fetch := func(key string) string {
		t.Helper()
		obj, err := store.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get %q: %v", key, err)
		}
		body, err := io.ReadAll(obj.Body)
		_ = obj.Body.Close()
		if err != nil {
			t.Fatalf("reading %q: %v", key, err)
		}
		return string(body)
	}

	// Single presigned upload with a content-type constraint.
	prefix := fmt.Sprintf("upload/%d", time.Now().UnixNano())
	key := prefix + "/single.txt"
	url, err := store.PresignedPutURL(ctx, key, "text/plain", 0)
	if err != nil {
		t.Fatalf("PresignedPutURL returned error: %v", err)
	}
	if !strings.Contains(url, "X-Amz-Signature=") {
		t.Fatalf("expected signed URL, got %q", url)
	}
	put(url, "text/plain", []byte("via presigned put"))
	if got := fetch(key); got != "via presigned put" {
		t.Fatalf("uploaded content mismatch: %q", got)
	}

	// Batch presigned uploads.
	bodies := map[string]string{
		prefix + "/a.txt": "batch-one",
		prefix + "/b.txt": "batch-two",
	}
	keys := []string{prefix + "/a.txt", prefix + "/b.txt"}
	urls, err := store.PresignedPutURLs(ctx, keys, "text/plain", 0)
	if err != nil {
		t.Fatalf("PresignedPutURLs returned error: %v", err)
	}
	if len(urls) != len(keys) {
		t.Fatalf("expected %d URLs, got %d", len(keys), len(urls))
	}
	for key, want := range bodies {
		put(urls[key], "text/plain", []byte(want))
		if got := fetch(key); got != want {
			t.Fatalf("uploaded content mismatch for %q: %q", key, got)
		}
		if err := store.Delete(ctx, key); err != nil {
			t.Fatalf("Delete %q: %v", key, err)
		}
	}
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete %q: %v", key, err)
	}
}
