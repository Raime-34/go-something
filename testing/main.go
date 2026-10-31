package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type appServer struct {
	shortener *URLShortener
}

func newServer(shortener *URLShortener) http.Handler {
	s := &appServer{shortener: shortener}
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", s.handleShorten)
	mux.HandleFunc("/", s.handleRedirect)
	return mux
}

func main() {
	addr := ":8080"
	log.Printf("сервер запущен на %s", addr)
	if err := http.ListenAndServe(addr, newServer(NewURLShortener())); err != nil {
		log.Fatal(err)
	}
}

// handleShorten обслуживает POST /shorten.
func (s *appServer) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	req, err := decodeShortenRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	shortID, err := s.shortener.Shorten(req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный URL")
		return
	}

	writeJSON(w, http.StatusOK, shortenResponse{
		ShortURL:    shortID,
		OriginalURL: req.URL,
	})
}

// handleRedirect обслуживает GET /{short_url} и отвечает 302 или 404.
func (s *appServer) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/")
	if shortID == "" || strings.Contains(shortID, "/") {
		writeError(w, http.StatusNotFound, "короткий URL не найден")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	original, err := s.shortener.GetOriginal(shortID)
	if err != nil {
		writeError(w, http.StatusNotFound, "короткий URL не найден")
		return
	}
	http.Redirect(w, r, original, http.StatusFound)
}

func decodeShortenRequest(r *http.Request) (shortenRequest, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req shortenRequest
	if err := dec.Decode(&req); err != nil {
		return shortenRequest{}, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return shortenRequest{}, errors.New("лишние данные в JSON")
	}
	return req, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
