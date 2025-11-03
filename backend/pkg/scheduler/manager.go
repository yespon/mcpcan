package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// DefaultTaskManager default task manager implementation
type DefaultTaskManager struct {
	scheduler Scheduler                // scheduler
	taskFuncs map[string]TaskFunc      // registered task functions
	mu        sync.RWMutex             // read-write lock
}

// NewTaskManager creates a new task manager
func NewTaskManager(scheduler Scheduler) *DefaultTaskManager {
	return &DefaultTaskManager{
		scheduler: scheduler,
		taskFuncs: make(map[string]TaskFunc),
	}
}

// RegisterTaskFunc registers a task function
func (tm *DefaultTaskManager) RegisterTaskFunc(name string, fn TaskFunc) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if name == "" {
		return fmt.Errorf("task function name cannot be empty")
	}

	if fn == nil {
		return fmt.Errorf("task function cannot be nil")
	}

	if _, exists := tm.taskFuncs[name]; exists {
		return fmt.Errorf("task function %s already exists", name)
	}

	tm.taskFuncs[name] = fn
	return nil
}

// GetTaskFunc gets a task function
func (tm *DefaultTaskManager) GetTaskFunc(name string) (TaskFunc, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	fn, exists := tm.taskFuncs[name]
	if !exists {
		return nil, fmt.Errorf("task function %s does not exist", name)
	}

	return fn, nil
}

// CreateCronTask creates a Cron task
func (tm *DefaultTaskManager) CreateCronTask(id, name, cronExpr, funcName string) (Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("task name cannot be empty")
	}

	if cronExpr == "" {
		return nil, fmt.Errorf("cron expression cannot be empty")
	}

	if funcName == "" {
		return nil, fmt.Errorf("task function name cannot be empty")
	}

	// Get task function
	taskFunc, err := tm.GetTaskFunc(funcName)
	if err != nil {
		return nil, fmt.Errorf("failed to get task function: %w", err)
	}

	// Create Cron task
	task, err := NewCronTask(id, name, cronExpr, funcName, taskFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron task: %w", err)
	}

	// Add to scheduler
	err = tm.scheduler.AddTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to add task to scheduler: %w", err)
	}

	return task, nil
}

// CreateTimerTask creates a timer task
func (tm *DefaultTaskManager) CreateTimerTask(id, name string, executeAt time.Time, funcName string) (Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("task name cannot be empty")
	}

	if funcName == "" {
		return nil, fmt.Errorf("task function name cannot be empty")
	}

	if executeAt.Before(time.Now()) {
		return nil, fmt.Errorf("execution time cannot be earlier than current time")
	}

	// Get task function
	taskFunc, err := tm.GetTaskFunc(funcName)
	if err != nil {
		return nil, fmt.Errorf("failed to get task function: %w", err)
	}

	// Create timer task
	task := NewTimerTask(id, name, executeAt, funcName, taskFunc)

	// Add to scheduler
	err = tm.scheduler.AddTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to add task to scheduler: %w", err)
	}

	return task, nil
}

// GetScheduler gets the scheduler
func (tm *DefaultTaskManager) GetScheduler() Scheduler {
	return tm.scheduler
}

// ListTaskFuncs lists all registered task functions
func (tm *DefaultTaskManager) ListTaskFuncs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	funcs := make([]string, 0, len(tm.taskFuncs))
	for name := range tm.taskFuncs {
		funcs = append(funcs, name)
	}

	return funcs
}

// RemoveTaskFunc removes a task function
func (tm *DefaultTaskManager) RemoveTaskFunc(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.taskFuncs[name]; !exists {
		return fmt.Errorf("task function %s does not exist", name)
	}

	delete(tm.taskFuncs, name)
	return nil
}

// GetTaskCount gets the number of tasks
func (tm *DefaultTaskManager) GetTaskCount() int {
	tasks := tm.scheduler.ListTasks()
	return len(tasks)
}

// GetTaskFuncCount gets the number of registered task functions
func (tm *DefaultTaskManager) GetTaskFuncCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.taskFuncs)
}

// IsSchedulerRunning checks if the scheduler is running
func (tm *DefaultTaskManager) IsSchedulerRunning() bool {
	return tm.scheduler.IsRunning()
}