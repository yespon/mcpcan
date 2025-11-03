package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// BaseTask base task struct
type BaseTask struct {
	id          string       // Task ID
	name        string       // Task name
	taskType    TaskType     // Task type
	status      TaskStatus   // Task status
	funcName    string       // Task function name
	taskFunc    TaskFunc     // Task execution function
	createdAt   time.Time    // Creation time
	lastRunTime *time.Time   // Last run time
	nextRunTime *time.Time   // Next run time
	mu          sync.RWMutex // Read-write lock
}

// GetID gets task ID
func (t *BaseTask) GetID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

// GetName gets task name
func (t *BaseTask) GetName() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.name
}

// GetType gets task type
func (t *BaseTask) GetType() TaskType {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.taskType
}

// GetStatus gets task status
func (t *BaseTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetNextRunTime gets next run time
func (t *BaseTask) GetNextRunTime() *time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nextRunTime
}

// GetLastRunTime gets last run time
func (t *BaseTask) GetLastRunTime() *time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastRunTime
}

// GetCreatedAt gets creation time
func (t *BaseTask) GetCreatedAt() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.createdAt
}

// setStatus sets task status (internal method)
func (t *BaseTask) setStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

// setLastRunTime sets last run time (internal method)
func (t *BaseTask) setLastRunTime(runTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRunTime = &runTime
}

// setNextRunTime sets next run time (internal method)
func (t *BaseTask) setNextRunTime(nextTime *time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextRunTime = nextTime
}

// Execute executes task
func (t *BaseTask) Execute(ctx context.Context) error {
	if t.taskFunc == nil {
		return fmt.Errorf("task function not set")
	}

	t.setStatus(TaskStatusRunning)
	t.setLastRunTime(time.Now())

	err := t.taskFunc(ctx)
	if err != nil {
		t.setStatus(TaskStatusFailed)
		return fmt.Errorf("task execution failed: %w", err)
	}

	t.setStatus(TaskStatusCompleted)
	return nil
}

// Cancel cancels task
func (t *BaseTask) Cancel() error {
	t.setStatus(TaskStatusCancelled)
	return nil
}

// CronTask Cron scheduled task
type CronTask struct {
	*BaseTask
	cronExpr string        // Cron expression
	schedule cron.Schedule // Cron scheduler
	entryID  cron.EntryID  // Cron entry ID
}

// NewCronTask creates new Cron task
func NewCronTask(id, name, cronExpr, funcName string, taskFunc TaskFunc) (*CronTask, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid Cron expression: %w", err)
	}

	now := time.Now()
	nextTime := schedule.Next(now)

	task := &CronTask{
		BaseTask: &BaseTask{
			id:          id,
			name:        name,
			taskType:    TaskTypeCron,
			status:      TaskStatusPending,
			funcName:    funcName,
			taskFunc:    taskFunc,
			createdAt:   now,
			nextRunTime: &nextTime,
		},
		cronExpr: cronExpr,
		schedule: schedule,
	}

	return task, nil
}

// GetCronExpr gets Cron expression
func (ct *CronTask) GetCronExpr() string {
	return ct.cronExpr
}

// UpdateNextRunTime updates next run time
func (ct *CronTask) UpdateNextRunTime() {
	now := time.Now()
	nextTime := ct.schedule.Next(now)
	ct.setNextRunTime(&nextTime)
}

// TimerTask one-time timer task
type TimerTask struct {
	*BaseTask
	executeAt time.Time   // Execution time
	timer     *time.Timer // Timer
}

// NewTimerTask creates new timer task
func NewTimerTask(id, name string, executeAt time.Time, funcName string, taskFunc TaskFunc) *TimerTask {
	now := time.Now()
	task := &TimerTask{
		BaseTask: &BaseTask{
			id:          id,
			name:        name,
			taskType:    TaskTypeTimer,
			status:      TaskStatusPending,
			funcName:    funcName,
			taskFunc:    taskFunc,
			createdAt:   now,
			nextRunTime: &executeAt,
		},
		executeAt: executeAt,
	}

	return task
}

// GetExecuteAt gets execution time
func (tt *TimerTask) GetExecuteAt() time.Time {
	return tt.executeAt
}

// Cancel cancels timer task
func (tt *TimerTask) Cancel() error {
	if tt.timer != nil {
		tt.timer.Stop()
	}
	return tt.BaseTask.Cancel()
}

// SetTimer sets timer
func (tt *TimerTask) SetTimer(timer *time.Timer) {
	tt.timer = timer
}
