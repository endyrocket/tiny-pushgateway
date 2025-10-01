// server_test.go
package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/* -------------------------------------------------------------------- */
/* 1. End‑to‑end push → scrape → scrape‑empty                           */
/* -------------------------------------------------------------------- */

func TestEndToEnd(t *testing.T) {
	data = nil // clear global buffer

	srv := httptest.NewServer(router())
	defer srv.Close()

	sample := `my_metric{label="test"} 1
`

	// push
	if resp, err := http.Post(srv.URL+"/push", "text/plain", strings.NewReader(sample)); err != nil || resp.StatusCode != http.StatusAccepted {
		t.Fatalf("push failed: %v status %d", err, resp.StatusCode)
	}

	// first scrape should return the sample
	body, _ := readAll(http.Get(srv.URL + "/metrics"))
	if string(body) != sample {
		t.Fatalf("want %q got %q", sample, body)
	}

	// second scrape should be empty
	body, _ = readAll(http.Get(srv.URL + "/metrics"))
	if len(body) != 0 {
		t.Fatalf("expected empty body on 2nd scrape, got %q", body)
	}
}

/* -------------------------------------------------------------------- */
/* 2. quickValidate helper                                              */
/* -------------------------------------------------------------------- */

func TestQuickValidate(t *testing.T) {
	valid := []byte("good_metric 1\nother_metric_total 2\n")
	if !quickValidate(valid) {
		t.Fatalf("valid exposition flagged false")
	}

	// line with explicit timestamp must be rejected
	if quickValidate([]byte("metric 1 123456\n")) {
		t.Fatalf("timestamp line not rejected")
	}

	// metric name starting with digit must be rejected
	if quickValidate([]byte("9bad 1\n")) {
		t.Fatalf("bad name not rejected")
	}
}

/* -------------------------------------------------------------------- */
/* 3. pushHandler rejects bad exposition                                */
/* -------------------------------------------------------------------- */

func TestPushRejectsInvalid(t *testing.T) {
	data = nil
	srv := httptest.NewServer(router())
	defer srv.Close()

	bad := "metric 1 123456\n" // contains timestamp -> invalid
	resp, _ := http.Post(srv.URL+"/push", "text/plain", strings.NewReader(bad))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

/* -------------------------------------------------------------------- */
/* 4. buffer drops after maxRetry failed scrapes                        */
/* -------------------------------------------------------------------- */

// failWriter always errors to simulate a failed scrape
type failWriter struct{ h http.Header }

func (fw *failWriter) Header() http.Header       { return fw.h }
func (fw *failWriter) WriteHeader(int)           {}
func (fw *failWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestMaxRetryDrop(t *testing.T) {
	data = []byte("metric 1\n")
	failCount = 0

	for i := 0; i < maxRetry; i++ {
		metricsHandler(&failWriter{h: make(http.Header)}, nil)
	}

	if len(data) != 0 {
		t.Fatalf("buffer not cleared after %d failed scrapes", maxRetry)
	}
	if failCount != 0 {
		t.Fatalf("failCount not reset, got %d", failCount)
	}
}

/* -------------------------------------------------------------------- */
/* Helpers                                                              */
/* -------------------------------------------------------------------- */

func router() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/push":
			pushHandler(w, r)
		case "/metrics":
			metricsHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

func readAll(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
			return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
