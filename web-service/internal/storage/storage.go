package storage

import (
	"errors"

	"tasks-api/internal/models"
)

var ErrNotFound = errors.New("task not found")

// Storage — хранилище задач. Реализация должна быть потокобезопасной.
type Storage interface {
	List() []models.Task
	Create(task models.Task) (models.Task, error)
	Get(id int) (models.Task, bool)
	Update(id int, task models.Task) (models.Task, error)
	Delete(id int) error
}
