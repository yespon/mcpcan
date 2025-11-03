package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskScheduler task scheduler implementation
type TaskScheduler struct {
	cronScheduler  *cron.Cron            // Cron scheduler
	timerTasks     map[string]*TimerTask // Timer task mapping
	tasks          map[string]Task       // All task mapping
	running        bool                  // Running status
	mu             sync.RWMutex          // Read-write lock
	ctx            context.Context       // Context
	cancel         context.CancelFunc    // Cancel function
	taskRepository TaskRepository        // Task repository
}

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler(taskRepository TaskRepository) *TaskScheduler {
	return &TaskScheduler{
		cronScheduler:  cron.New(cron.WithSeconds()),
		timerTasks:     make(map[string]*TimerTask),
		tasks:          make(map[string]Task),
		running:        false,
		taskRepository: taskRepository,
	}
}

// Start starts the scheduler
func (s *TaskScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	// Start Cron scheduler
	s.cronScheduler.Start()

	// Start timer task monitoring
	go s.monitorTimerTasks()

	return nil
}

// Stop stops the scheduler
func (s *TaskScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	// Stop Cron scheduler
	s.cronScheduler.Stop()

	// Cancel all timer tasks
	for _, task := range s.timerTasks {
		task.Cancel()
	}

	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	return nil
}

// AddTask adds a task
func (s *TaskScheduler) AddTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskID := task.GetID()
	if _, exists := s.tasks[taskID]; exists {
		return fmt.Errorf("task ID %s already exists", taskID)
	}

	// Add to corresponding scheduler based on task type
	switch t := task.(type) {
	case *CronTask:
		err := s.addCronTask(t)
		if err != nil {
			return err
		}
	case *TimerTask:
		err := s.addTimerTask(t)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported task type")
	}

	s.tasks[taskID] = task

	// Save task to repository
	if s.taskRepository != nil {
		if err := s.taskRepository.SaveTask(s.ctx, task); err != nil {
			return fmt.Errorf("failed to save task: %w", err)
		}
	}

	return nil
}

// RemoveTask removes a task
func (s *TaskScheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task ID %s does not exist", taskID)
	}

	// Remove from corresponding scheduler based on task type
	switch t := task.(type) {
	case *CronTask:
		s.cronScheduler.Remove(t.entryID)
	case *TimerTask:
		t.Cancel()
		delete(s.timerTasks, taskID)
	}

	delete(s.tasks, taskID)

	// Delete task from repository
	if s.taskRepository != nil {
		if err := s.taskRepository.DeleteTask(taskID); err != nil {
			return fmt.Errorf("failed to delete task: %w", err)
		}
	}

	return nil
}

// GetTask gets a task
func (s *TaskScheduler) GetTask(taskID string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task ID %s does not exist", taskID)
	}

	return task, nil
}

// ListTasks lists all tasks
func (s *TaskScheduler) ListTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// IsRunning checks if the scheduler is running
func (s *TaskScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// addCronTask adds a Cron task
func (s *TaskScheduler) addCronTask(task *CronTask) error {
	entryID, err := s.cronScheduler.AddFunc(task.GetCronExpr(), func() {
		s.executeCronTask(task)
	})
	if err != nil {
		return fmt.Errorf("failed to add Cron task: %w", err)
	}

	task.entryID = entryID
	return nil
}

// addTimerTask adds a timer task
func (s *TaskScheduler) addTimerTask(task *TimerTask) error {
	now := time.Now()
	executeAt := task.GetExecuteAt()

	if executeAt.Before(now) {
		return fmt.Errorf("execution time cannot be earlier than current time")
	}

	duration := executeAt.Sub(now)
	timer := time.NewTimer(duration)
	task.SetTimer(timer)

	s.timerTasks[task.GetID()] = task

	return nil
}

// executeCronTask executes a Cron task
func (s *TaskScheduler) executeCronTask(task *CronTask) {
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Minute) // 60 minute timeout
		defer cancel()

		err := task.Execute(ctx)
		if err != nil {
			// Log error
			fmt.Printf("Cron task execution failed [%s]: %v\n", task.GetName(), err)
		}

		// Update next run time
		task.UpdateNextRunTime()

		// Update task status to repository
		if s.taskRepository != nil {
			s.taskRepository.UpdateTaskStatus(task.GetID(), task.GetStatus())
		}
	}()
}

// monitorTimerTasks monitors timer tasks
func (s *TaskScheduler) monitorTimerTasks() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			s.checkTimerTasks()
			time.Sleep(1 * time.Second) // Check every second
		}
	}
}

// checkTimerTasks checks timer tasks
func (s *TaskScheduler) checkTimerTasks() {
	s.mu.RLock()
	tasks := make([]*TimerTask, 0, len(s.timerTasks))
	for _, task := range s.timerTasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		select {
		case <-task.timer.C:
			s.executeTimerTask(task)
		default:
			// Task has not reached execution time yet
		}
	}
}

// executeTimerTask executes a timer task
func (s *TaskScheduler) executeTimerTask(task *TimerTask) {
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Minute) // 60 minute timeout
		defer cancel()

		err := task.Execute(ctx)
		if err != nil {
			// Log error
			fmt.Printf("Timer task execution failed [%s]: %v\n", task.GetName(), err)
		}

		// Update task status to repository
		if s.taskRepository != nil {
			s.taskRepository.UpdateTaskStatus(task.GetID(), task.GetStatus())
		}

		// Remove completed timer task
		s.mu.Lock()
		delete(s.timerTasks, task.GetID())
		delete(s.tasks, task.GetID())
		s.mu.Unlock()
	}()
}
