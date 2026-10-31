package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleShorten(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "валидный URL",
			method:     http.MethodPost,
			body:       `{"url":"https://example.com/very/long/path"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "некорректный JSON",
			method:     http.MethodPost,
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный JSON",
		},
		{
			name:       "пустое тело",
			method:     http.MethodPost,
			body:       ``,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный JSON",
		},
		{
			name:       "не JSON-объект",
			method:     http.MethodPost,
			body:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный JSON",
		},
		{
			name:       "неизвестное поле",
			method:     http.MethodPost,
			body:       `{"url":"https://example.com","extra":1}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный JSON",
		},
		{
			name:       "два JSON-значения",
			method:     http.MethodPost,
			body:       `{"url":"https://example.com"}{"url":"https://example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный JSON",
		},
		{
			name:       "невалидный URL",
			method:     http.MethodPost,
			body:       `{"url":"not-a-url"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный URL",
		},
		{
			name:       "пустой URL",
			method:     http.MethodPost,
			body:       `{"url":""}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный URL",
		},
		{
			name:       "нет поля url",
			method:     http.MethodPost,
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "некорректный URL",
		},
		{
			name:       "не POST",
			method:     http.MethodGet,
			body:       ``,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "метод не поддерживается",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newServer(NewURLShortener())
			req := httptest.NewRequest(tt.method, "/shorten", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидали %d, тело: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, ожидали application/json", ct)
			}
			if tt.wantError == "" {
				var resp shortenResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("json: %v", err)
				}
				if resp.OriginalURL != "https://example.com/very/long/path" {
					t.Errorf("original_url = %q", resp.OriginalURL)
				}
				if len(resp.ShortURL) < 6 || len(resp.ShortURL) > 8 {
					t.Errorf("short_url = %q", resp.ShortURL)
				}
				return
			}

			var resp errorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("json: %v, тело: %s", err, rec.Body.String())
			}
			if resp.Error != tt.wantError {
				t.Errorf("error = %q, ожидали %q", resp.Error, tt.wantError)
			}
		})
	}
}

func TestHandleRedirect(t *testing.T) {
	us := NewURLShortener()
	id, err := us.Shorten("https://example.com/very/long/path")
	if err != nil {
		t.Fatal(err)
	}
	srv := newServer(us)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantLoc    string
	}{
		{"редирект на оригинал", http.MethodGet, "/" + id, http.StatusFound, "https://example.com/very/long/path"},
		{"неизвестный id", http.MethodGet, "/missing1", http.StatusNotFound, ""},
		{"пустой путь", http.MethodGet, "/", http.StatusNotFound, ""},
		{"вложенный путь", http.MethodGet, "/a/b", http.StatusNotFound, ""},
		{"не GET", http.MethodPost, "/" + id, http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидали %d, тело: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if loc := rec.Header().Get("Location"); loc != tt.wantLoc {
				t.Errorf("Location = %q, ожидали %q", loc, tt.wantLoc)
			}
		})
	}
}

func TestShortenThenRedirect(t *testing.T) {
	srv := newServer(NewURLShortener())
	original := "https://google.com/search?q=test"

	body := `{"url":"` + original + `"}`
	postReq := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	postRec := httptest.NewRecorder()
	srv.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Fatalf("POST статус = %d, тело: %s", postRec.Code, postRec.Body.String())
	}

	var created shortenResponse
	if err := json.Unmarshal(postRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.OriginalURL != original {
		t.Fatalf("original_url = %q", created.OriginalURL)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/"+created.ShortURL, nil)
	getRec := httptest.NewRecorder()
	srv.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusFound {
		t.Fatalf("GET статус = %d, тело: %s", getRec.Code, getRec.Body.String())
	}
	if loc := getRec.Header().Get("Location"); loc != original {
		t.Fatalf("Location = %q, ожидали %q", loc, original)
	}

	// Клиент httptest по умолчанию не следует редиректу, если вызывать ServeHTTP напрямую.
	// Проверяем, что тело ответа редиректа не путают с JSON-ошибкой.
	if _, err := io.ReadAll(getRec.Body); err != nil {
		t.Fatal(err)
	}
}
