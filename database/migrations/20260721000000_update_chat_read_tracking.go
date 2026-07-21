package migrations

import (
	"github.com/Hieu3z03/chat-api-golang/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(
		"20260721000000_update_chat_read_tracking",
		Up20260721000000UpdateChatReadTracking,
		Down20260721000000UpdateChatReadTracking,
	)
}

func Up20260721000000UpdateChatReadTracking(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS name varchar(200)",
			"UPDATE users SET name = COALESCE(NULLIF(btrim(concat_ws(' ', first_name, last_name)), ''), username) WHERE name IS NULL OR btrim(name) = ''",
			"ALTER TABLE users ALTER COLUMN name SET NOT NULL",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url varchar(2048)",
			"ALTER TABLE users DROP COLUMN IF EXISTS first_name",
			"ALTER TABLE users DROP COLUMN IF EXISTS last_name",
			"ALTER TABLE users DROP COLUMN IF EXISTS avatar_id",
			"ALTER TABLE users DROP COLUMN IF EXISTS created_at",
			"ALTER TABLE users DROP COLUMN IF EXISTS updated_at",
			"ALTER TABLE channel_members ADD COLUMN IF NOT EXISTS last_read_sequence bigint NOT NULL DEFAULT 0",
			"ALTER TABLE channel_members ADD COLUMN IF NOT EXISTS last_read_at timestamp with time zone",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func Down20260721000000UpdateChatReadTracking(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name varchar(100)",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name varchar(100)",
			"UPDATE users SET first_name = COALESCE(NULLIF(split_part(btrim(name), ' ', 1), ''), username), last_name = CASE WHEN position(' ' in btrim(name)) > 0 THEN btrim(substr(btrim(name), position(' ' in btrim(name)) + 1)) ELSE '' END",
			"ALTER TABLE users ALTER COLUMN first_name SET NOT NULL",
			"ALTER TABLE users ALTER COLUMN last_name SET NOT NULL",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_id uuid",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at timestamp with time zone",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone",
			"ALTER TABLE users DROP COLUMN IF EXISTS name",
			"ALTER TABLE users DROP COLUMN IF EXISTS avatar_url",
			"ALTER TABLE channel_members DROP COLUMN IF EXISTS last_read_at",
			"ALTER TABLE channel_members DROP COLUMN IF EXISTS last_read_sequence",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
