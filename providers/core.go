package providers

import (
	"github.com/Hieu3z03/chat-api-golang/config"
	authController "github.com/Hieu3z03/chat-api-golang/modules/auth/controller"
	authRepo "github.com/Hieu3z03/chat-api-golang/modules/auth/repository"
	authService "github.com/Hieu3z03/chat-api-golang/modules/auth/service"
	userController "github.com/Hieu3z03/chat-api-golang/modules/user/controller"
	"github.com/Hieu3z03/chat-api-golang/modules/user/repository"
	userService "github.com/Hieu3z03/chat-api-golang/modules/user/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/constants"
	"github.com/samber/do"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

func InitDatabases(injector *do.Injector) {
	do.Provide(injector, func(i *do.Injector) (*config.PostgreSQLConnection, error) {
		return config.SetUpDatabaseConnection()
	})
	do.ProvideNamed(injector, constants.PostgreSQL, func(i *do.Injector) (*gorm.DB, error) {
		connection, err := do.Invoke[*config.PostgreSQLConnection](i)
		if err != nil {
			return nil, err
		}
		return connection.DB, nil
	})

	do.Provide(injector, func(i *do.Injector) (*config.MongoDBConnection, error) {
		return config.SetUpMongoDBConnection()
	})
	do.ProvideNamed(injector, constants.MongoDB, func(i *do.Injector) (*mongo.Database, error) {
		connection, err := do.Invoke[*config.MongoDBConnection](i)
		if err != nil {
			return nil, err
		}
		return connection.Database, nil
	})
}

func RegisterDependencies(injector *do.Injector) error {
	InitDatabases(injector)

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db, err := do.InvokeNamed[*gorm.DB](injector, constants.PostgreSQL)
	if err != nil {
		return err
	}
	if _, err := do.InvokeNamed[*mongo.Database](injector, constants.MongoDB); err != nil {
		return err
	}

	jwtService, err := do.InvokeNamed[authService.JWTService](injector, constants.JWTService)
	if err != nil {
		return err
	}

	userRepository := repository.NewUserRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)

	userService := userService.NewUserService(userRepository, db)
	authService := authService.NewAuthService(userRepository, refreshTokenRepository, jwtService, db)

	do.Provide(
		injector, func(i *do.Injector) (userController.UserController, error) {
			return userController.NewUserController(i, userService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (authController.AuthController, error) {
			return authController.NewAuthController(i, authService), nil
		},
	)

	return nil
}
