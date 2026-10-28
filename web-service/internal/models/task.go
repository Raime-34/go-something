package models

// Task — задача в списке to-do.
// ID и CreatedAt выставляет сервер, клиент их не задаёт.
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at,omitempty"`
}
