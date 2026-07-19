package migrations

import (
	"github.com/Hieu3z03/chat-api-golang/database"
	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(
		"20260719010000_create_chat_schema",
		Up20260719010000CreateChatSchema,
		Down20260719010000CreateChatSchema,
	)
}

func Up20260719010000CreateChatSchema(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"DROP TABLE IF EXISTS refresh_tokens",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS username varchar(100)",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_id uuid",
			"UPDATE users SET username = 'user_' || substring(replace(id::text, '-', '') from 1 for 12) WHERE username IS NULL OR btrim(username) = ''",
			"ALTER TABLE users ALTER COLUMN username SET NOT NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username)",
			"ALTER TABLE users DROP COLUMN IF EXISTS email",
			"ALTER TABLE users DROP COLUMN IF EXISTS telp_number",
			"ALTER TABLE users DROP COLUMN IF EXISTS password",
			"ALTER TABLE users DROP COLUMN IF EXISTS role",
			"ALTER TABLE users DROP COLUMN IF EXISTS image_url",
			"ALTER TABLE users DROP COLUMN IF EXISTS is_verified",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}

		return tx.AutoMigrate(
			&entities.User{},
			&entities.Channel{},
			&entities.ChannelMember{},
		)
	})
}

func Down20260719010000CreateChatSchema(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&entities.ChannelMember{},
		&entities.Channel{},
	)
}
