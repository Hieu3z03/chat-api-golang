package repository

import (
	"context"
	"time"

	"github.com/Hieu3z03/chat-api-golang/modules/message/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MessageRepository interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, message model.Message) (model.Message, error)
	ListByChannel(ctx context.Context, channelID string, limit int64, before *time.Time) ([]model.Message, error)
}

type messageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository(database *mongo.Database) MessageRepository {
	return &messageRepository{collection: database.Collection("messages")}
}

func (repository *messageRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "channel_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	})
	return err
}

func (repository *messageRepository) Create(ctx context.Context, message model.Message) (model.Message, error) {
	message.ID = bson.NewObjectID()
	message.CreatedAt = time.Now().UTC()

	if _, err := repository.collection.InsertOne(ctx, message); err != nil {
		return model.Message{}, err
	}

	return message, nil
}

func (repository *messageRepository) ListByChannel(
	ctx context.Context,
	channelID string,
	limit int64,
	before *time.Time,
) ([]model.Message, error) {
	filter := bson.D{{Key: "channel_id", Value: channelID}}
	if before != nil {
		filter = append(filter, bson.E{
			Key:   "created_at",
			Value: bson.D{{Key: "$lt", Value: before.UTC()}},
		})
	}

	cursor, err := repository.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetLimit(limit),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []model.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
