package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/Hieu3z03/chat-api-golang/config"
	channelController "github.com/Hieu3z03/chat-api-golang/modules/channel/controller"
	channelRepository "github.com/Hieu3z03/chat-api-golang/modules/channel/repository"
	channelService "github.com/Hieu3z03/chat-api-golang/modules/channel/service"
	messageController "github.com/Hieu3z03/chat-api-golang/modules/message/controller"
	messageRepository "github.com/Hieu3z03/chat-api-golang/modules/message/repository"
	messageService "github.com/Hieu3z03/chat-api-golang/modules/message/service"
	realtimeController "github.com/Hieu3z03/chat-api-golang/modules/realtime/controller"
	realtimeService "github.com/Hieu3z03/chat-api-golang/modules/realtime/service"
	userController "github.com/Hieu3z03/chat-api-golang/modules/user/controller"
	userRepository "github.com/Hieu3z03/chat-api-golang/modules/user/repository"
	userService "github.com/Hieu3z03/chat-api-golang/modules/user/service"
	"github.com/Hieu3z03/chat-api-golang/pkg/centrifugo"
	"github.com/Hieu3z03/chat-api-golang/pkg/constants"
	"github.com/redis/go-redis/v9"
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

	do.Provide(injector, func(i *do.Injector) (*config.RedisConnection, error) {
		return config.SetUpRedisConnection()
	})

	do.ProvideNamed(injector, constants.Redis, func(i *do.Injector) (*redis.Client, error) {
		connection, err := do.Invoke[*config.RedisConnection](i)
		if err != nil {
			return nil, err
		}
		return connection.Client, nil
	})
}

func RegisterDependencies(injector *do.Injector) error {
	InitDatabases(injector)

	postgres, err := do.InvokeNamed[*gorm.DB](injector, constants.PostgreSQL)
	if err != nil {
		return err
	}
	mongodb, err := do.InvokeNamed[*mongo.Database](injector, constants.MongoDB)
	if err != nil {
		return err
	}

	users := userRepository.NewUserRepository(postgres)
	channels := channelRepository.NewChannelRepository(postgres)
	messages := messageRepository.NewMessageRepository(mongodb)

	indexContext, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := messages.EnsureIndexes(indexContext); err != nil {
		return fmt.Errorf("ensure MongoDB message indexes: %w", err)
	}

	userUseCases := userService.NewUserService(users)
	channelUseCases := channelService.NewChannelService(channels, users)
	centrifugoSettings := config.LoadCentrifugoSettings()

	centrifugoClient := centrifugo.NewClient(
		centrifugoSettings.APIURL,
		centrifugoSettings.APIKey,
		centrifugoSettings.TokenHMACSecret,
	)

	realtimeUseCases := realtimeService.NewRealtimeService(
		centrifugoClient,
		channels,
	)
	centrifugoContext, centrifugoCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer centrifugoCancel()
	if err := centrifugoClient.Ping(centrifugoContext); err != nil {
		return fmt.Errorf("check Centrifugo connection: %w", err)
	}
	messageUseCases := messageService.NewMessageService(messages, channelUseCases, centrifugoClient)
	realtimeService.NewRealtimeService(
		centrifugoClient, channels,
	)

	do.Provide(injector, func(i *do.Injector) (userController.UserController, error) {
		return userController.NewUserController(userUseCases), nil
	})
	do.Provide(injector, func(i *do.Injector) (channelController.ChannelController, error) {
		return channelController.NewChannelController(channelUseCases), nil
	})
	do.Provide(injector, func(i *do.Injector) (messageController.MessageController, error) {
		return messageController.NewMessageController(messageUseCases), nil
	})
	do.Provide(injector, func(i *do.Injector) (realtimeController.RealtimeController, error) {
		return realtimeController.NewRealtimeController(realtimeUseCases), nil
	})

	return nil
}
