package script

import (
	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"gorm.io/gorm"
)

type (
	ExampleScript struct {
		db *gorm.DB
	}
)

func NewExampleScript(db *gorm.DB) *ExampleScript {
	return &ExampleScript{
		db: db,
	}
}

func (s *ExampleScript) Run() error {
	// your script here
	appLogger.Log(nil).Info().Str("component", "script").Msg("example script running")
	return nil
}
