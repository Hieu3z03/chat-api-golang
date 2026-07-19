package migrations

import (
	"github.com/Hieu3z03/chat-api-golang/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(
		"20260719020000_split_user_name",
		Up20260719020000SplitUserName,
		Down20260719020000SplitUserName,
	)
}

func Up20260719020000SplitUserName(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name varchar(100)",
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name varchar(100)",
			`DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'users'
			AND column_name = 'name'
	) THEN
		UPDATE users
		SET first_name = COALESCE(NULLIF(split_part(btrim(name), ' ', 1), ''), username),
			last_name = CASE
				WHEN position(' ' in btrim(name)) > 0
				THEN btrim(substr(btrim(name), position(' ' in btrim(name)) + 1))
				ELSE ''
			END
		WHERE first_name IS NULL OR last_name IS NULL;
	END IF;
END $$`,
			"UPDATE users SET first_name = COALESCE(NULLIF(first_name, ''), username), last_name = COALESCE(last_name, '')",
			"ALTER TABLE users ALTER COLUMN first_name SET NOT NULL",
			"ALTER TABLE users ALTER COLUMN last_name SET NOT NULL",
			"ALTER TABLE users DROP COLUMN IF EXISTS name",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func Down20260719020000SplitUserName(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"ALTER TABLE users ADD COLUMN IF NOT EXISTS name varchar(201)",
			"UPDATE users SET name = btrim(first_name || ' ' || last_name)",
			"ALTER TABLE users ALTER COLUMN name SET NOT NULL",
			"ALTER TABLE users DROP COLUMN IF EXISTS first_name",
			"ALTER TABLE users DROP COLUMN IF EXISTS last_name",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
