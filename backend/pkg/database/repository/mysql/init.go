package mysql

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	once   sync.Once
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	conf   *Config
)

// Config defines MySQL connection configuration.
type Config struct {
	Host                string        `validate:"required"`
	Port                int           `validate:"required,min=1,max=65535"`
	Username            string        `validate:"required"`
	Password            string        `validate:"required"`
	Database            string        `validate:"required"`
	ConnectTimeout      time.Duration `validate:"required"`
	MaxIdleConns        int           `validate:"min=0"`
	MaxOpenConns        int           `validate:"min=0"`
	HealthCheckInterval time.Duration `validate:"required"`
	MaxRetries          int           `validate:"min=0"`
	RetryInterval       time.Duration `validate:"required"`
}

// InitHook represents a function to be called after initial DB setup.
type InitHook func()

// HookManager manages registered initialization hooks.
type HookManager struct {
	hooks []InitHook
	mu    sync.RWMutex
}

var hookManager = NewHookManager()

// NewHookManager creates a new HookManager.
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]InitHook, 0),
	}
}

// Register adds an initialization hook to the manager.
func (m *HookManager) Register(hook InitHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// CallHooks executes all registered initialization hooks.
func (m *HookManager) CallHooks() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, hook := range m.hooks {
		hook()
	}
}

// RegisterInit registers an initialization hook for DB setup.
func RegisterInit(initHook InitHook) {
	hookManager.Register(initHook)
}

// InitDB initializes the database connection once and starts health checking.
func InitDB(config *Config) error {
	if config == nil {
		return errors.New("no mysql config")
	}

	var initErr error
	once.Do(func() {
		// Create context with cancel for health checker.
		ctx, cancel = context.WithCancel(context.Background())

		// Initialize DB connection.
		if err := initConnection(config); err != nil {
			initErr = err
			return
		}

		// Start health checker and automatic reconnect goroutine.
		startHealthChecker(config.HealthCheckInterval, config.MaxRetries, config.RetryInterval)
	})

	return initErr
}

// buildDSN builds a DSN string using provided Config.
func buildDSN(config *Config) string {
	t := config.ConnectTimeout.String()
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		t,
		t,
		t,
	)
}

// initConnection initializes the database connection using the given config.
func initConnection(config *Config) error {
	dsn := buildDSN(config)

	var err error
	db, err = gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	if healthErr := HealthCheck(); healthErr != nil {
		return fmt.Errorf("failed to health check: %v", healthErr)
	}

	// Set connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// Set pool parameters.
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour * 1)
	sqlDB.SetConnMaxIdleTime(time.Minute * 5)

	// Persist config for future reconnect attempts.
	mu.Lock()
	conf = config
	mu.Unlock()

	// Call all initialization hooks (only on initial setup).
	hookManager.CallHooks()

	return nil
}

// startHealthChecker starts periodic health checks and triggers reconnect when unhealthy.
func startHealthChecker(interval time.Duration, maxRetries int, retryInterval time.Duration) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := HealthCheck(); err != nil {
					fmt.Printf("Database health check failed: %v\n", err)
					if err := reconnect(maxRetries, retryInterval); err != nil {
						fmt.Printf("Failed to reconnect to database: %v\n", err)
					} else {
						fmt.Println("Successfully reconnected to database")
					}
				}
			}
		}
	}()
}

// reconnect attempts to rebuild the DB connection and swap it atomically.
func reconnect(maxRetries int, retryInterval time.Duration) error {
	var lastErr error
	var cfg *Config

	mu.RLock()
	cfg = conf
	mu.RUnlock()

	if cfg == nil {
		return errors.New("missing mysql config for reconnect")
	}

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Retrying to connect to database (attempt %d/%d)...\n", i, maxRetries)
			time.Sleep(retryInterval)
		}

		// Create a new DB connection using the stored config.
		newDB, err := gorm.Open(
			mysql.Open(buildDSN(cfg)),
			&gorm.Config{
				Logger:                                   logger.Default.LogMode(logger.Info),
				DisableForeignKeyConstraintWhenMigrating: true,
			},
		)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect: %v", err)
			continue
		}

		newSQLDB, err := newDB.DB()
		if err != nil {
			lastErr = fmt.Errorf("failed to get database instance: %v", err)
			continue
		}

		// Apply pool configuration.
		newSQLDB.SetMaxIdleConns(cfg.MaxIdleConns)
		newSQLDB.SetMaxOpenConns(cfg.MaxOpenConns)

		// Validate the new connection.
		if err := newSQLDB.Ping(); err != nil {
			lastErr = fmt.Errorf("failed to ping: %v", err)
			_ = newSQLDB.Close()
			continue
		}

		// Swap global DB atomically.
		mu.Lock()
		old := db
		db = newDB
		mu.Unlock()

		// Close the old connection pool after swap.
		if old != nil {
			if oldSQL, err := old.DB(); err == nil {
				_ = oldSQL.Close()
			}
		}

		// Re-run initialization hooks to ensure schema and indexes exist after reconnect.
		hookManager.CallHooks()

		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %v", maxRetries, lastErr)
}

// GetDB returns the global DB instance, ensuring it is healthy.
func GetDB() *gorm.DB {
	// Attempt auto-heal if unhealthy.
	if err := HealthCheck(); err != nil {
		mu.RLock()
		cfg := conf
		mu.RUnlock()
		if cfg != nil {
			_ = reconnect(cfg.MaxRetries, cfg.RetryInterval)
		}
	}

	mu.RLock()
	defer mu.RUnlock()
	return db
}

// HealthCheck performs a ping on the underlying sql.DB.
func HealthCheck() error {
	mu.RLock()
	defer mu.RUnlock()

	if db == nil {
		return errors.New("database connection is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	return sqlDB.Ping()
}

// Close stops health checking and closes the DB connection pool.
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	// Cancel health checker goroutine.
	if cancel != nil {
		cancel()
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	return sqlDB.Close()
}
