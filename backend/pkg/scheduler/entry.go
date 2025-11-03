package scheduler

import (
	"context"
	"sync"
	"time"
)

// GlobalSchedulerEntry global scheduler entry
type GlobalSchedulerEntry struct {
	ctx         context.Context
	taskManager TaskManager
	once        sync.Once
}

var (
	// globalEntry global scheduler instance
	globalEntry *GlobalSchedulerEntry
	// globalOnce ensures global instance is initialized only once
	globalOnce sync.Once
)

// GetGlobalScheduler gets global scheduler instance
func GetGlobalScheduler() *GlobalSchedulerEntry {
	globalOnce.Do(func() {
		// Create memory task repository (can be replaced with database storage)
		taskRepo := NewMemoryTaskRepository()

		// Create scheduler
		scheduler := NewTaskScheduler(taskRepo)

		// Create task manager
		taskManager := NewTaskManager(scheduler)

		globalEntry = &GlobalSchedulerEntry{
			ctx:         context.Background(),
			taskManager: taskManager,
		}
	})
	return globalEntry
}

// Start starts global scheduler
func (g *GlobalSchedulerEntry) Start() error {
	return g.taskManager.GetScheduler().Start(g.ctx)
}

// Stop stops global scheduler
func (g *GlobalSchedulerEntry) Stop() error {
	return g.taskManager.GetScheduler().Stop()
}

// GetTaskManager gets task manager
func (g *GlobalSchedulerEntry) GetTaskManager() TaskManager {
	return g.taskManager
}

// RegisterTaskFunc registers task function
func (g *GlobalSchedulerEntry) RegisterTaskFunc(name string, fn TaskFunc) error {
	return g.taskManager.RegisterTaskFunc(name, fn)
}

// Convenience functions, directly use global instance

// Start starts global scheduler (convenience function)
func Start(ctx context.Context) error {
	return GetGlobalScheduler().Start()
}

// Stop stops global scheduler (convenience function)
func Stop() error {
	return GetGlobalScheduler().Stop()
}

// RegisterTaskFunc registers task function (convenience function)
func RegisterTaskFunc(name string, fn TaskFunc) error {
	return GetGlobalScheduler().RegisterTaskFunc(name, fn)
}

// GetTaskManager gets task manager (convenience function)
func GetTaskManager() TaskManager {
	return GetGlobalScheduler().GetTaskManager()
}

// CreateCronTask creates Cron task (convenience function)
func CreateCronTask(id, name, cronExpr, funcName string) (Task, error) {
	return GetGlobalScheduler().GetTaskManager().CreateCronTask(id, name, cronExpr, funcName)
}

// CreateTimerTask creates timer task (convenience function)
func CreateTimerTask(id, name string, executeAt time.Time, funcName string) (Task, error) {
	return GetGlobalScheduler().GetTaskManager().CreateTimerTask(id, name, executeAt, funcName)
}

// GetTask gets task (convenience function)
func GetTask(taskID string) (Task, error) {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().GetTask(taskID)
}

// RemoveTask removes task (convenience function)
func RemoveTask(taskID string) error {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().RemoveTask(taskID)
}

// ListTasks lists all tasks (convenience function)
func ListTasks() []Task {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().ListTasks()
}

// IsRunning checks if scheduler is running (convenience function)
func IsRunning() bool {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().IsRunning()
}
