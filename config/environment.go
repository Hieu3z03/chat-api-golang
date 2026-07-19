package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvironment() error {
	err := godotenv.Load(".env")
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return fmt.Errorf("load .env: %w", err)
}
