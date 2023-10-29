package database

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	dbInstanceSqlite   *gorm.DB
	dbInstancePostgres *gorm.DB
	sqliteMutex        sync.Mutex
	postgresMutex      sync.Mutex
)

type File struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

func InitSqliteDbInstance() (*gorm.DB, error) {
	sqliteMutex.Lock()
	defer sqliteMutex.Unlock()

	if dbInstanceSqlite == nil {
		db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		migrationPath := "migrations/file.sql"
		if err := runMigrations(db, migrationPath); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %v", err)
		}

		dbInstanceSqlite = db
	}

	if dbInstanceSqlite == nil {
		return nil, fmt.Errorf("failed to initialize SQLite DB instance")
	}
	return dbInstanceSqlite, nil
}
func InitPostgresDbInstance() (*gorm.DB, error) {
	postgresMutex.Lock()
	defer postgresMutex.Unlock()

	if dbInstancePostgres == nil {
		db, err := gorm.Open(postgres.Open("host=localhost user=myuser dbname=mydatabase password=mypassword port=5432 sslmode=disable"), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		migrationPath := "migrations/file.sql"
		if err := runMigrations(db, migrationPath); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %v", err)
		}
		sqlDB, errDb := db.DB()
		if errDb != nil {
			return nil, errDb
		}

		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Minute)

		dbInstancePostgres = db
	}

	if dbInstancePostgres == nil {
		return nil, fmt.Errorf("failed to initialize Postgres database instance")
	}
	return dbInstancePostgres, nil
}

func runMigrations(db *gorm.DB, filePath string) error {
	migrations, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Execute the migrations with gorm
	if err = db.Exec(string(migrations)).Error; err != nil {
		return err
	}

	return nil
}
