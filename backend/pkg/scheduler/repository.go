package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/kymo-mcp/mcpcan/pkg/i18n"
)

// MemoryTaskRepository memory task repository implementation
type MemoryTaskRepository struct {
	tasks map[string]Task // Task mapping
	mu    sync.RWMutex    // Read-write lock
}

// NewMemoryTaskRepository creates a new memory task repository
func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[string]Task),
	}
}

// SaveTask saves a task
func (r *MemoryTaskRepository) SaveTask(ctx context.Context, task Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task == nil {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeTaskCannotBeEmpty))
	}

	taskID := task.GetID()
	if taskID == "" {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeTaskIDCannotBeEmpty))
	}

	r.tasks[taskID] = task
	return nil
}

// GetTask gets a task
func (r *MemoryTaskRepository) GetTask(taskID string) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	task, exists := r.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task ID %s does not exist", taskID)
	}

	return task, nil
}

// ListTasks lists tasks
func (r *MemoryTaskRepository) ListTasks() ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// DeleteTask deletes a task
func (r *MemoryTaskRepository) DeleteTask(taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	if _, exists := r.tasks[taskID]; !exists {
		return fmt.Errorf("task ID %s does not exist", taskID)
	}

	delete(r.tasks, taskID)
	return nil
}

// UpdateTaskStatus updates task status
func (r *MemoryTaskRepository) UpdateTaskStatus(taskID string, status TaskStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	task, exists := r.tasks[taskID]
	if !exists {
		return fmt.Errorf("task ID %s does not exist", taskID)
	}

	// 由于Task接口没有SetStatus方法，这里只能通过类型断言来更新状态
	switch t := task.(type) {
	case *CronTask:
		t.setStatus(status)
	case *TimerTask:
		t.setStatus(status)
	default:
		return fmt.Errorf("unsupported task type: %T", task)
	}

	return nil
}

// GetTaskCount gets task count
func (r *MemoryTaskRepository) GetTaskCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tasks)
}

// Clear clears all tasks
func (r *MemoryTaskRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = make(map[string]Task)
	return nil
}
