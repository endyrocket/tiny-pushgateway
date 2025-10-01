package main

import (
	"io"
	"log"
	"net/http"
	"sync"
	"bufio"
	"bytes"
	"unicode"
	"unicode/utf8"
)

var (
	mu   sync.Mutex
	data []byte
)
var failCount int
const maxRetry = 3 // drop buffer after 3 consecutive failures

// isValidMetricStart reports whether r is a legal first rune of a metric name.
func isValidMetricStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// quickValidate returns true if every non‑comment/non‑blank line:
//   • begins with a legal metric identifier rune
//   • contains 1 or 2 whitespace‑separated fields (name+value or name{…}+value)
func quickValidate(b []byte) bool {
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		// fast check: first rune must be [A‑Za‑z_:] or _
		r, _ := utf8.DecodeRune(line)
		if !isValidMetricStart(r) {
			return false
		}
		// count fields (split on ASCII space)
		if fields := bytes.Count(line, []byte(" ")); fields > 1 {
			return false
		}
	}
	return sc.Err() == nil
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	if !quickValidate(body) {
		http.Error(w, "invalid exposition format", http.StatusBadRequest)
		return
	}

	mu.Lock()
	data = append(data, body...)
	mu.Unlock()

	w.WriteHeader(http.StatusAccepted)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if len(data) == 0 {
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if _, err := w.Write(data); err != nil {
		failCount++
		if failCount >= maxRetry {
			log.Printf("dropping buffer after %d failed scrapes", failCount)
			data = data[:0]
			failCount = 0
		}
		return // keep buffer until we hit limit
	}

	// scrape succeeded
	data = data[:0]
	failCount = 0
}

func main() {
	http.HandleFunc("/push", pushHandler)
	http.HandleFunc("/metrics", metricsHandler)

	addr := ":9091"
	log.Printf("listening on %s …", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
