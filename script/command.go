package script

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/Hieu3z03/chat-api-golang/database/migrations"

	"github.com/Hieu3z03/chat-api-golang/database"
	"github.com/Hieu3z03/chat-api-golang/pkg/constants"
	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func Commands(injector *do.Injector) (bool, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.PostgreSQL)

	var scriptName string
	var migrationName string
	var rollbackBatch int

	migrateRun := false
	migrateRollback := false
	migrateRollbackAll := false
	migrateStatus := false
	migrateCreate := false
	seed := false
	run := false
	scriptFlag := false

	for i, arg := range os.Args[1:] {
		if arg == "--migrate" || arg == "--migrate:run" {
			migrateRun = true
		}
		if arg == "--migrate:rollback" {
			migrateRollback = true
			if i+2 < len(os.Args) && !strings.HasPrefix(os.Args[i+2], "--") {
				batch, err := strconv.Atoi(os.Args[i+2])
				if err == nil {
					rollbackBatch = batch
				}
			}
		}
		if arg == "--migrate:rollback:all" {
			migrateRollbackAll = true
		}
		if arg == "--migrate:status" {
			migrateStatus = true
		}
		if strings.HasPrefix(arg, "--migrate:create:") {
			migrateCreate = true
			migrationName = strings.TrimPrefix(arg, "--migrate:create:")
		}
		if arg == "--seed" {
			seed = true
		}
		if arg == "--run" {
			run = true
		}
		if strings.HasPrefix(arg, "--script:") {
			scriptFlag = true
			scriptName = strings.TrimPrefix(arg, "--script:")
		}
	}

	if migrateRun {
		if err := database.Migrate(db); err != nil {
			return false, fmt.Errorf("run migration: %w", err)
		}
		logCommandSuccess("migrate")
	}

	if migrateRollback {
		manager := database.NewMigrationManager(db)
		if rollbackBatch > 0 {
			if err := manager.Rollback(rollbackBatch); err != nil {
				return false, fmt.Errorf("rollback migration batch %d: %w", rollbackBatch, err)
			}
		} else {
			if err := manager.Rollback(0); err != nil {
				return false, fmt.Errorf("rollback migration: %w", err)
			}
		}
		logCommandSuccess("migrate:rollback")
	}

	if migrateRollbackAll {
		manager := database.NewMigrationManager(db)
		if err := manager.RollbackAll(); err != nil {
			return false, fmt.Errorf("rollback all migrations: %w", err)
		}
		logCommandSuccess("migrate:rollback:all")
	}

	if migrateStatus {
		manager := database.NewMigrationManager(db)
		if err := manager.Status(); err != nil {
			return false, fmt.Errorf("get migration status: %w", err)
		}
	}

	if migrateCreate {
		if migrationName == "" {
			return false, errors.New("migration name is required")
		}
		manager := database.NewMigrationManager(db)
		if err := manager.Create(migrationName); err != nil {
			return false, fmt.Errorf("create migration: %w", err)
		}
		logCommandSuccess("migrate:create")
	}

	if seed {
		if err := database.Seeder(db); err != nil {
			return false, fmt.Errorf("run seeder: %w", err)
		}
		logCommandSuccess("seed")
	}

	if scriptFlag {
		if err := Script(scriptName, db); err != nil {
			return false, fmt.Errorf("run script: %w", err)
		}
		logCommandSuccess("script:" + scriptName)
	}

	if run {
		return true, nil
	}

	if migrateRun || migrateRollback || migrateRollbackAll || migrateStatus || migrateCreate {
		return false, nil
	}

	return false, nil
}

func logCommandSuccess(command string) {
	appLogger.Log(nil).Info().
		Str("component", "command").
		Str("command", command).
		Msg("command completed")
}
