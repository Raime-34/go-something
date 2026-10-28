package storage

import (
	"sort"
	"sync"
	"time"

	"tasks-api/internal/models"
)

// Memory хранит задачи в памяти процесса.
type Memory struct {
	mu     sync.RWMutex
	nextID int
	tasks  map[int]models.Task
}

func NewMemory() *Memory {
	return &Memory{
		nextID: 1,
		tasks:  make(map[int]models.Task),
	}
}

func (m *Memory) List() []models.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		out = append(out, task)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (m *Memory) Create(task models.Task) (models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task.ID = m.nextID
	m.nextID++
	if task.CreatedAt == "" {
		task.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *Memory) Get(id int) (models.Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	return task, ok
}

func (m *Memory) Update(id int, task models.Task) (models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	existing.Title = task.Title
	existing.Done = task.Done
	m.tasks[id] = existing
	return existing, nil
}

func (m *Memory) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}
