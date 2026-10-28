package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"tasks-api/internal/models"
	"tasks-api/internal/storage"
)

type Handler struct {
	Store storage.Storage
}

func New(s storage.Storage) *Handler {
	return &Handler{Store: s}
}

type errorBody struct {
	Error string `json:"error"`
}

type taskInput struct {
	ID        *int    `json:"id"`
	Title     *string `json:"title"`
	Done      *bool   `json:"done"`
	CreatedAt *string `json:"created_at"`
}

// TasksCollection обслуживает /tasks: GET — список, POST — создание.
func (h *Handler) TasksCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}

// TaskItem обслуживает /tasks/{id}: GET, PUT, DELETE.
func (h *Handler) TaskItem(w http.ResponseWriter, r *http.Request) {
	idPart := strings.Trim(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/")
	if idPart == "" {
		h.TasksCollection(w, r)
		return
	}
	if strings.Contains(idPart, "/") {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}

	id, err := strconv.Atoi(idPart)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, id)
	case http.MethodPut:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}

// Health — проверка, что процесс жив.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listTasks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.Store.List())
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	title, done, ok := readTask(w, r, false)
	if !ok {
		return
	}

	created, err := h.Store.Create(models.Task{Title: title, Done: done})
	if err != nil {
		log.Printf("create task: %v", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) getTask(w http.ResponseWriter, id int) {
	task, found := h.Store.Get(id)
	if !found {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, id int) {
	title, done, ok := readTask(w, r, true)
	if !ok {
		return
	}

	updated, err := h.Store.Update(id, models.Task{Title: title, Done: done})
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		log.Printf("update task: %v", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteTask(w http.ResponseWriter, id int) {
	err := h.Store.Delete(id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		log.Printf("delete task: %v", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func readTask(w http.ResponseWriter, r *http.Request, requireDone bool) (string, bool, bool) {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		writeError(w, http.StatusBadRequest, "Content-Type должен быть application/json")
		return "", false, false
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var in taskInput
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return "", false, false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return "", false, false
	}

	if in.Title == nil || strings.TrimSpace(*in.Title) == "" {
		writeError(w, http.StatusBadRequest, "поле title обязательно")
		return "", false, false
	}
	if requireDone && in.Done == nil {
		writeError(w, http.StatusBadRequest, "поле done обязательно")
		return "", false, false
	}

	done := false
	if in.Done != nil {
		done = *in.Done
	}
	return strings.TrimSpace(*in.Title), done, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}
