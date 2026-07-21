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
	counters   *mongo.Collection
}

func NewMessageRepository(database *mongo.Database) MessageRepository {
	return &messageRepository{
		collection: database.Collection("messages"),
		counters:   database.Collection("message_counters"),
	}
}

func (repository *messageRepository) EnsureIndexes(ctx context.Context) error {
	if err := repository.backfillSequences(ctx); err != nil {
		return err
	}

	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "channel_id", Value: 1},
				{Key: "sequence", Value: -1},
			},
			Options: options.Index().
				SetName("idx_messages_channel_sequence").
				SetUnique(true).
				SetPartialFilterExpression(bson.D{{
					Key:   "sequence",
					Value: bson.D{{Key: "$gt", Value: 0}},
				}}),
		},
		{
			Keys: bson.D{
				{Key: "channel_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	})
	return err
}

func (repository *messageRepository) Create(ctx context.Context, message model.Message) (model.Message, error) {
	sequence, err := repository.nextSequence(ctx, message.ChannelID)
	if err != nil {
		return model.Message{}, err
	}

	message.ID = bson.NewObjectID()
	message.Sequence = sequence
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
			SetSort(bson.D{
				{Key: "sequence", Value: -1},
				{Key: "created_at", Value: -1},
			}).
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

func (repository *messageRepository) nextSequence(ctx context.Context, channelID string) (int64, error) {
	var counter struct {
		Sequence int64 `bson:"sequence"`
	}
	err := repository.counters.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: channelID}},
		bson.D{{Key: "$inc", Value: bson.D{{Key: "sequence", Value: 1}}}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.Sequence, nil
}

func (repository *messageRepository) backfillSequences(ctx context.Context) error {
	var channelIDs []string
	if err := repository.collection.Distinct(ctx, "channel_id", bson.D{}).Decode(&channelIDs); err != nil {
		return err
	}

	for _, channelID := range channelIDs {
		maxSequence, err := repository.maxSequence(ctx, channelID)
		if err != nil {
			return err
		}
		if err := repository.raiseCounter(ctx, channelID, maxSequence); err != nil {
			return err
		}

		missingFilter := bson.D{
			{Key: "channel_id", Value: channelID},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "sequence", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "sequence", Value: nil}},
				bson.D{{Key: "sequence", Value: bson.D{{Key: "$lte", Value: 0}}}},
			}},
		}
		cursor, err := repository.collection.Find(
			ctx,
			missingFilter,
			options.Find().
				SetProjection(bson.D{{Key: "_id", Value: 1}}).
				SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "_id", Value: 1}}),
		)
		if err != nil {
			return err
		}

		var messages []struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err := cursor.All(ctx, &messages); err != nil {
			cursor.Close(ctx)
			return err
		}
		cursor.Close(ctx)

		writes := make([]mongo.WriteModel, 0, len(messages))
		for _, message := range messages {
			maxSequence++
			writes = append(writes, mongo.NewUpdateOneModel().
				SetFilter(bson.D{{Key: "_id", Value: message.ID}}).
				SetUpdate(bson.D{{Key: "$set", Value: bson.D{{Key: "sequence", Value: maxSequence}}}}))
		}
		if len(writes) > 0 {
			if _, err := repository.collection.BulkWrite(ctx, writes); err != nil {
				return err
			}
			if err := repository.raiseCounter(ctx, channelID, maxSequence); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repository *messageRepository) maxSequence(ctx context.Context, channelID string) (int64, error) {
	var message struct {
		Sequence int64 `bson:"sequence"`
	}
	err := repository.collection.FindOne(
		ctx,
		bson.D{
			{Key: "channel_id", Value: channelID},
			{Key: "sequence", Value: bson.D{{Key: "$gt", Value: 0}}},
		},
		options.FindOne().
			SetProjection(bson.D{{Key: "sequence", Value: 1}}).
			SetSort(bson.D{{Key: "sequence", Value: -1}}),
	).Decode(&message)
	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return message.Sequence, nil
}

func (repository *messageRepository) raiseCounter(ctx context.Context, channelID string, sequence int64) error {
	_, err := repository.counters.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: channelID}},
		bson.D{{Key: "$max", Value: bson.D{{Key: "sequence", Value: sequence}}}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}
