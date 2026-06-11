package database

import (
	"freegfw/models"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(path string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			ParameterizedQueries:      true,        // Do not print parameter values in logs
			Colorful:                  false,       // Disable colorful printing
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound to reduce noisy logs
		},
	)

	// DSN parameters explanation:
	//   _journal_mode=WAL   — WAL mode enables concurrent read/write (default DELETE locks the whole file during writes)
	//   _busy_timeout=5000  — Wait up to 5 seconds when encountering lock contention instead of returning SQLITE_BUSY immediately
	//                         This is crucial to resolve freezes/errors caused by concurrent writes from multiple goroutines
	//   _synchronous=NORMAL — NORMAL is safe enough in WAL mode and performs better than FULL
	//   _cache_size=-8000   — Page cache size 8MB (negative value indicates size in KB)
	//   _foreign_keys=ON    — Enable foreign key constraints
	dsn := path + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=-8000&_foreign_keys=ON"

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// SQLite uses file-level locking, allowing only one write transaction at a time.
	// Setting the connection pool size to 1 prevents multiple connections from competing for the write lock,
	// and allows _busy_timeout to work properly (WAL is required when busy_timeout is effective across multiple connections in the same process).
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	sqlDB.SetMaxOpenConns(1)              // SQLite single writer, 1 connection is sufficient
	sqlDB.SetMaxIdleConns(1)              // Keep 1 idle connection to avoid frequent recreation
	sqlDB.SetConnMaxLifetime(time.Hour)   // Connection maximum lifetime 1 hour
	sqlDB.SetConnMaxIdleTime(time.Minute) // Release connection if idle for more than 1 minute

	err = DB.AutoMigrate(&models.User{}, &models.Link{}, &models.Setting{}, &models.Template{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
