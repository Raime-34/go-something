package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
)

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"валидный HTTP URL", "http://example.com", false},
		{"валидный HTTPS URL", "https://google.com/search?q=test", false},
		{"URL с путём", "https://example.com/very/long/path", false},
		{"URL с фрагментом", "https://example.com/docs#section", false},
		{"localhost", "http://localhost:8080/health", false},
		{"IPv4", "https://127.0.0.1/path", false},
		{"схема в верхнем регистре", "HTTP://example.com/Path", false},
		{"невалидный URL", "not-a-url", true},
		{"пустая строка", "", true},
		{"без схемы", "example.com", true},
		{"не http схема", "ftp://example.com/file", true},
		{"пустой host", "http://", true},
		{"пробел в адресе", "https://example.com/a b", true},
		{"относительный путь", "/local/path", true},
	}

	shortener := NewURLShortener()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := shortener.Shorten(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Errorf("ошибка = %v, ожидали ErrInvalidURL", err)
				}
				return
			}
			if len(shortID) < 6 || len(shortID) > 8 {
				t.Errorf("длина короткого ID = %d, ожидали 6–8: %s", len(shortID), shortID)
			}
			got, err := shortener.GetOriginal(shortID)
			if err != nil {
				t.Fatalf("GetOriginal() error = %v", err)
			}
			if got != tt.url {
				t.Errorf("GetOriginal() = %q, ожидали %q", got, tt.url)
			}
		})
	}
}

func TestURLShortener_GetOriginal(t *testing.T) {
	us := NewURLShortener()
	id, err := us.Shorten("https://example.com/path")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		want    string
		wantErr error
	}{
		{"существующий id", id, "https://example.com/path", nil},
		{"несуществующий id", "missing", "", ErrNotFound},
		{"пустой id", "", "", ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := us.GetOriginal(tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ошибка = %v, ожидали %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("url = %q, ожидали %q", got, tt.want)
			}
		})
	}
}

func TestURLShortener_UniqueIDs(t *testing.T) {
	us := NewURLShortener()
	seen := make(map[string]struct{})
	for i := 0; i < 30; i++ {
		id, err := us.Shorten(fmt.Sprintf("https://example.com/item/%d", i))
		if err != nil {
			t.Fatalf("Shorten() error = %v", err)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("идентификатор %s повторился", id)
		}
		seen[id] = struct{}{}
	}
}

func TestURLShortener_Concurrent(t *testing.T) {
	us := NewURLShortener()
	const n = 40

	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]error, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ids[i], errs[i] = us.Shorten(fmt.Sprintf("https://example.com/c/%d", i))
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("Shorten(%d) error = %v", i, errs[i])
		}
		got, err := us.GetOriginal(ids[i])
		if err != nil {
			t.Fatalf("GetOriginal(%q) error = %v", ids[i], err)
		}
		want := fmt.Sprintf("https://example.com/c/%d", i)
		if got != want {
			t.Fatalf("GetOriginal(%q) = %q, ожидали %q", ids[i], got, want)
		}
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{"http", "http://example.com", true},
		{"https с query", "https://google.com/search?q=test", true},
		{"пусто", "", false},
		{"без схемы", "example.com", false},
		{"ftp", "ftp://example.com", false},
		{"javascript", "javascript:alert(1)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.raw); got != tt.want {
				t.Errorf("isValidURL(%q) = %v, ожидали %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestShorten_IDCollision(t *testing.T) {
	first := []byte{1, 2, 3, 4, 5, 6}
	second := []byte{7, 8, 9, 10, 11, 12}
	taken := base64.RawURLEncoding.EncodeToString(first)

	orig := rand.Reader
	rand.Reader = &scriptedReader{data: append(append([]byte{}, first...), second...)}
	t.Cleanup(func() { rand.Reader = orig })

	us := NewURLShortener()
	us.mu.Lock()
	us.urls[taken] = "https://example.com/taken"
	us.mu.Unlock()

	id, err := us.Shorten("https://example.com/new")
	if err != nil {
		t.Fatal(err)
	}
	want := base64.RawURLEncoding.EncodeToString(second)
	if id != want {
		t.Fatalf("id = %q, ожидали %q", id, want)
	}

	got, err := us.GetOriginal(taken)
	if err != nil || got != "https://example.com/taken" {
		t.Fatalf("прежняя запись = %q, %v", got, err)
	}
	got, err = us.GetOriginal(id)
	if err != nil || got != "https://example.com/new" {
		t.Fatalf("новая запись = %q, %v", got, err)
	}
}

type scriptedReader struct {
	data []byte
	off  int
}

func (s *scriptedReader) Read(p []byte) (int, error) {
	if s.off >= len(s.data) {
		return 0, io.EOF
	}
	n := copy(p, s.data[s.off:])
	s.off += n
	return n, nil
}

func TestGenerateShortID(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 20; i++ {
		id := generateShortID()
		if len(id) < 6 || len(id) > 8 {
			t.Fatalf("длина = %d, ожидали 6–8: %s", len(id), id)
		}
		for _, r := range id {
			ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
			if !ok {
				t.Fatalf("недопустимый символ %q в %s", r, id)
			}
		}
		seen[id] = struct{}{}
	}
	if len(seen) != 20 {
		t.Fatalf("ожидали 20 уникальных id, получили %d", len(seen))
	}
}
