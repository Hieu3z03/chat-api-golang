package main

import (
	"log"
	"os"

	"github.com/Hieu3z03/chat-api-golang/config"
	"github.com/Hieu3z03/chat-api-golang/middlewares"
	"github.com/Hieu3z03/chat-api-golang/modules/auth"
	"github.com/Hieu3z03/chat-api-golang/modules/user"
	"github.com/Hieu3z03/chat-api-golang/providers"
	"github.com/Hieu3z03/chat-api-golang/script"
	"github.com/samber/do"

	"github.com/gin-gonic/gin"
)

func args(injector *do.Injector) bool {
	if len(os.Args) > 1 {
		flag := script.Commands(injector)
		return flag
	}

	return true
}

func run(server *gin.Engine) error {
	server.Static("/assets", "./assets")

	port := os.Getenv("GOLANG_PORT")
	if port == "" {
		port = "8888"
	}

	return server.Run(":" + port)
}

func main() {
	if err := config.LoadEnvironment(); err != nil {
		log.Fatal(err)
	}

	injector := do.New()
	defer func() {
		if err := injector.Shutdown(); err != nil {
			log.Printf("shutdown dependencies: %v", err)
		}
	}()

	if err := providers.RegisterDependencies(injector); err != nil {
		log.Printf("initialize dependencies: %v", err)
		return
	}

	if !args(injector) {
		return
	}

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

	// Register module routes
	user.RegisterRoutes(server, injector)
	auth.RegisterRoutes(server, injector)

	if err := run(server); err != nil {
		log.Printf("run server: %v", err)
	}
}
