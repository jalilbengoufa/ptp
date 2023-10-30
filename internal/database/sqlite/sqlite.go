package sqlite

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jalilbengoufa/ptp/internal/messaging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DbInstanceSqlite *gorm.DB
	sqliteMutex      sync.Mutex
)

type File struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

func PullDataFromSqlite() {

	for {
		time.Sleep(2 * time.Second) // Poll every second

		// Read from sqlite
		var fileText File
		DbInstanceSqlite.First(&fileText, "1")
		fmt.Println("Read from sqlite:", fileText.Content)

		// save to topic

		topic := "fileEdit"
		topicSaveError := messaging.Producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(fileText.Content),
		}, nil)
		if topicSaveError != nil {
			fmt.Printf("Failed to save to topic: %s\n", topicSaveError)
		}
	}

}
func InitSqliteDbInstance() (*gorm.DB, error) {
	sqliteMutex.Lock()
	defer sqliteMutex.Unlock()

	if DbInstanceSqlite == nil {
		db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		migrationPath := "migrations/file.sql"
		if err := runMigrations(db, migrationPath); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %v", err)
		}

		DbInstanceSqlite = db
	}

	if DbInstanceSqlite == nil {
		return nil, fmt.Errorf("failed to initialize SQLite DB instance")
	}
	return DbInstanceSqlite, nil
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
