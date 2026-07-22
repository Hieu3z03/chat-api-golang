package main

import (
	"os"

	"github.com/Hieu3z03/chat-api-golang/config"
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/channel"
	"github.com/Hieu3z03/chat-api-golang/modules/message"
	"github.com/Hieu3z03/chat-api-golang/modules/realtime"
	"github.com/Hieu3z03/chat-api-golang/modules/user"
	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"github.com/Hieu3z03/chat-api-golang/providers"
	"github.com/Hieu3z03/chat-api-golang/script"
	"github.com/samber/do"

	"github.com/gin-gonic/gin"
)

func args(injector *do.Injector) (bool, error) {
	if len(os.Args) > 1 {
		return script.Commands(injector)
	}

	return true, nil
}

func run(server *gin.Engine, port string) error {
	server.Static("/assets", "./assets")
	return server.Run(":" + port)
}

func serverPort() string {
	if port := os.Getenv("GOLANG_PORT"); port != "" {
		return port
	}
	return "8888"
}

func main() {
	if err := config.LoadEnvironment(); err != nil {
		appLogger.Log(nil).Error().Err(err).Msg("load environment")
		return
	}
	if err := appLogger.ConfigureFromEnv(); err != nil {
		appLogger.Log(nil).Error().Err(err).Msg("configure logger")
		return
	}

	injector := do.New()
	defer func() {
		if err := injector.Shutdown(); err != nil {
			appLogger.Log(nil).Error().Err(err).Msg("shutdown dependencies")
		}
	}()

	if err := providers.RegisterDependencies(injector); err != nil {
		appLogger.Log(nil).Error().Err(err).Msg("initialize dependencies")
		return
	}

	shouldRun, err := args(injector)
	if err != nil {
		appLogger.Log(nil).Error().Err(err).Msg("run command")
		return
	}
	if !shouldRun {
		return
	}

	server := gin.New()
	server.Use(middlewares.RequestID())
	server.Use(middlewares.HTTPLogger())
	server.Use(middlewares.ErrorLogger())
	server.Use(middlewares.Recovery())
	server.Use(middlewares.CORSMiddleware())

	// Register module routes
	user.RegisterRoutes(server, injector)
	channel.RegisterRoutes(server, injector)
	message.RegisterRoutes(server, injector)
	realtime.RegisterRoutes(server, injector)

	port := serverPort()
	appLogger.Log(nil).Info().
		Str("component", "startup").
		Str("server_port", port).
		Str("gin_mode", gin.Mode()).
		Bool("postgresql_connected", true).
		Bool("mongodb_connected", true).
		Bool("centrifugo_connected", true).
		Msg("server starting")

	if err := run(server, port); err != nil {
		appLogger.Log(nil).Error().Err(err).Msg("run server")
	}
}
