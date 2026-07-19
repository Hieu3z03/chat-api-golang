package repository

import (
	"context"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChannelRepository interface {
	Create(ctx context.Context, channel entities.Channel, memberIDs []uuid.UUID) (entities.Channel, error)
	FindByID(ctx context.Context, channelID uuid.UUID) (entities.Channel, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Channel, error)
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
}

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

func (repository *channelRepository) Create(
	ctx context.Context,
	channel entities.Channel,
	memberIDs []uuid.UUID,
) (entities.Channel, error) {
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&channel).Error; err != nil {
			return err
		}

		members := make([]entities.ChannelMember, 0, len(memberIDs))
		for _, userID := range memberIDs {
			members = append(members, entities.ChannelMember{
				ChannelID: channel.ID,
				UserID:    userID,
			})
		}

		return tx.Create(&members).Error
	})
	if err != nil {
		return entities.Channel{}, err
	}

	return repository.FindByID(ctx, channel.ID)
}

func (repository *channelRepository) FindByID(ctx context.Context, channelID uuid.UUID) (entities.Channel, error) {
	var channel entities.Channel
	err := repository.db.WithContext(ctx).
		Preload("Creator").
		Preload("Members.User").
		First(&channel, "id = ?", channelID).
		Error
	if err != nil {
		return entities.Channel{}, err
	}

	return channel, nil
}

func (repository *channelRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Channel, error) {
	var channels []entities.Channel
	err := repository.db.WithContext(ctx).
		Joins("JOIN channel_members ON channel_members.channel_id = channels.id").
		Where("channel_members.user_id = ?", userID).
		Preload("Creator").
		Preload("Members.User").
		Order("channels.updated_at DESC").
		Find(&channels).
		Error
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (repository *channelRepository) IsMember(
	ctx context.Context,
	channelID uuid.UUID,
	userID uuid.UUID,
) (bool, error) {
	var count int64
	err := repository.db.WithContext(ctx).
		Model(&entities.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count).
		Error
	return count > 0, err
}
