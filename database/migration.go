package database

import (
	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&entities.Migration{}); err != nil {
		return err
	}

	manager := NewMigrationManager(db)
	return manager.Run()
}
