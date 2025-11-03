package scheduler

import (
	"context"
	"time"
)

// TaskFunc task execution function type
type TaskFunc func(ctx context.Context) error

// TaskStatus task status
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // Pending
	TaskStatusRunning                     // Running
	TaskStatusCompleted                   // Completed
	TaskStatusFailed                      // Failed
	TaskStatusCancelled                   // Cancelled
)

// TaskType task type
type TaskType int

const (
	TaskTypeCron  TaskType = iota // Cron scheduled task
	TaskTypeTimer                 // One-time timer task
)

// Task task interface
type Task interface {
	// GetID gets task ID
	GetID() string
	// GetName gets task name
	GetName() string
	// GetType gets task type
	GetType() TaskType
	// GetStatus gets task status
	GetStatus() TaskStatus
	// Execute executes task
	Execute(ctx context.Context) error
	// Cancel cancels task
	Cancel() error
	// GetNextRunTime gets next run time
	GetNextRunTime() *time.Time
	// GetLastRunTime gets last run time
	GetLastRunTime() *time.Time
	// GetCreatedAt gets creation time
	GetCreatedAt() time.Time
}

// Scheduler scheduler interface
type Scheduler interface {
	// Start starts scheduler
	Start(ctx context.Context) error
	// Stop stops scheduler
	Stop() error
	// AddTask adds task
	AddTask(task Task) error
	// RemoveTask removes task
	RemoveTask(taskID string) error
	// GetTask gets task
	GetTask(taskID string) (Task, error)
	// ListTasks lists all tasks
	ListTasks() []Task
	// IsRunning checks if scheduler is running
	IsRunning() bool
}

// TaskManager task manager interface
type TaskManager interface {
	// RegisterTaskFunc registers task function
	RegisterTaskFunc(name string, fn TaskFunc) error
	// GetTaskFunc gets task function
	GetTaskFunc(name string) (TaskFunc, error)
	// CreateCronTask creates Cron task
	CreateCronTask(id, name, cronExpr, funcName string) (Task, error)
	// CreateTimerTask creates timer task
	CreateTimerTask(id, name string, executeAt time.Time, funcName string) (Task, error)
	// GetScheduler gets scheduler
	GetScheduler() Scheduler
}

// TaskRepository task repository interface
type TaskRepository interface {
	// SaveTask saves task
	SaveTask(ctx context.Context, task Task) error
	// GetTask gets task
	GetTask(taskID string) (Task, error)
	// ListTasks lists tasks
	ListTasks() ([]Task, error)
	// DeleteTask deletes task
	DeleteTask(taskID string) error
	// UpdateTaskStatus updates task status
	UpdateTaskStatus(taskID string, status TaskStatus) error
}
