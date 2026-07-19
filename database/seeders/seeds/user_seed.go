package seeds

import (
	"encoding/json"
	"io"
	"os"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ListUserSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/seeders/json/users.json")
	if err != nil {
		return err
	}

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listUser []entities.User
	if err := json.Unmarshal(jsonData, &listUser); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entities.User{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entities.User{}); err != nil {
			return err
		}
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"first_name",
			"last_name",
			"username",
			"avatar_id",
			"updated_at",
		}),
	}).Create(&listUser).Error
}
