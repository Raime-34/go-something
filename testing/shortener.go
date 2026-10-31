package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"sync"
)

var (
	// ErrInvalidURL — адрес не является корректным http/https URL.
	ErrInvalidURL = errors.New("некорректный URL")
	// ErrNotFound — короткий идентификатор отсутствует в хранилище.
	ErrNotFound = errors.New("короткий URL не найден")
)

// URLShortener хранит соответствие коротких идентификаторов и исходных адресов.
type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

// Shorten создает короткий идентификатор для URL.
func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", ErrInvalidURL
	}

	for {
		id := generateShortID()

		us.mu.Lock()
		if _, exists := us.urls[id]; exists {
			us.mu.Unlock()
			continue
		}
		us.urls[id] = originalURL
		us.mu.Unlock()
		return id, nil
	}
}

// GetOriginal возвращает оригинальный URL по короткому ID.
func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	original, ok := us.urls[shortID]
	if !ok {
		return "", ErrNotFound
	}
	return original, nil
}

// generateShortID генерирует случайный короткий идентификатор длиной 8 символов.
func generateShortID() string {
	buf := make([]byte, 6)
	// crypto/rand.Read заполняет буфер целиком и не возвращает ошибку.
	_, _ = rand.Read(buf)
	// 6 байт в base64 без padding — ровно 8 символов из алфавита URL-safe.
	return base64.RawURLEncoding.EncodeToString(buf)
}

// isValidURL проверяет, что строка — абсолютный http или https адрес.
func isValidURL(raw string) bool {
	if raw == "" || strings.ContainsAny(raw, " \t\r\n") {
		return false
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return false
	}
	return true
}
