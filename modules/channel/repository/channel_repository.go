package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChannelRepository interface {
	Create(ctx context.Context, channel entities.Channel, memberIDs []uuid.UUID) (entities.Channel, error)
	FindByID(ctx context.Context, channelID uuid.UUID) (entities.Channel, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Channel, error)
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	ListMembersAndMarkRead(
		ctx context.Context,
		channelID, userID uuid.UUID,
		lastReadSequence int64,
	) ([]ChannelMemberWithUser, error)
}

type ChannelMemberWithUser struct {
	ChannelID        uuid.UUID  `gorm:"column:channel_id"`
	UserID           uuid.UUID  `gorm:"column:user_id"`
	JoinedAt         time.Time  `gorm:"column:joined_at"`
	LastReadSequence int64      `gorm:"column:last_read_sequence"`
	LastReadAt       *time.Time `gorm:"column:last_read_at"`
	Username         string     `gorm:"column:username"`
	Name             string     `gorm:"column:name"`
	AvatarURL        *string    `gorm:"column:avatar_url"`
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
	if len(memberIDs) == 0 {
		return entities.Channel{}, gorm.ErrRecordNotFound
	}

	var duplicateErr error
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		memberSet := make(map[uuid.UUID]struct{}, len(memberIDs))
		for _, memberID := range memberIDs {
			memberSet[memberID] = struct{}{}
		}

		var existingMembers []entities.ChannelMember
		if err := tx.Model(&entities.ChannelMember{}).
			Select("channel_id").
			Group("channel_id").
			Having("COUNT(DISTINCT user_id) = ?", len(memberSet)).
			Scan(&existingMembers).Error; err != nil {
			return err
		}

		if len(existingMembers) > 0 {
			memberIDsByChannel := make(map[uuid.UUID][]uuid.UUID)
			for _, member := range existingMembers {
				memberIDsByChannel[member.ChannelID] = nil
			}

			for channelID := range memberIDsByChannel {
				var matchedMembers []entities.ChannelMember
				if err := tx.Where("channel_id = ?", channelID).Find(&matchedMembers).Error; err != nil {
					return err
				}
				matchedSet := make(map[uuid.UUID]struct{}, len(matchedMembers))
				for _, matchedMember := range matchedMembers {
					matchedSet[matchedMember.UserID] = struct{}{}
				}
				if len(matchedSet) == len(memberSet) {
					allMatch := true
					for userID := range memberSet {
						if _, exists := matchedSet[userID]; !exists {
							allMatch = false
							break
						}
					}
					if allMatch {
						duplicateErr = dto.ErrChannelAlreadyExists
						return nil
					}
				}
			}
		}

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
	if duplicateErr != nil {
		return entities.Channel{}, duplicateErr
	}
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

func (repository *channelRepository) ListMembersAndMarkRead(
	ctx context.Context,
	channelID, userID uuid.UUID,
	lastReadSequence int64,
) ([]ChannelMemberWithUser, error) {
	const query = `
WITH updated_member AS (
	UPDATE channel_members
	SET last_read_sequence = GREATEST(last_read_sequence, @last_read_sequence),
		last_read_at = CURRENT_TIMESTAMP
	WHERE channel_id = @channel_id AND user_id = @user_id
	RETURNING user_id, last_read_sequence, last_read_at
)
SELECT
	cm.channel_id,
	cm.user_id,
	cm.joined_at,
	COALESCE(updated_member.last_read_sequence, cm.last_read_sequence) AS last_read_sequence,
	COALESCE(updated_member.last_read_at, cm.last_read_at) AS last_read_at,
	u.username,
	u.name,
	u.avatar_url
FROM channel_members AS cm
JOIN users AS u ON u.id = cm.user_id
LEFT JOIN updated_member ON updated_member.user_id = cm.user_id
WHERE cm.channel_id = @channel_id
	AND EXISTS (
		SELECT 1
		FROM channel_members AS requester
		WHERE requester.channel_id = @channel_id AND requester.user_id = @user_id
	)
ORDER BY cm.joined_at ASC, cm.user_id ASC`

	var members []ChannelMemberWithUser
	err := repository.db.WithContext(ctx).Raw(
		query,
		sql.Named("channel_id", channelID),
		sql.Named("user_id", userID),
		sql.Named("last_read_sequence", lastReadSequence),
	).Scan(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}
