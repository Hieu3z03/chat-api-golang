package migrations

import (
	"github.com/Hieu3z03/chat-api-golang/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(
		"20260719000000_drop_is_verified_from_users_table",
		Up20260719000000DropIsVerifiedFromUsersTable,
		Down20260719000000DropIsVerifiedFromUsersTable,
	)
}

func Up20260719000000DropIsVerifiedFromUsersTable(db *gorm.DB) error {
	return db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS is_verified").Error
}

func Down20260719000000DropIsVerifiedFromUsersTable(db *gorm.DB) error {
	return db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified boolean DEFAULT false").Error
}
