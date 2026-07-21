package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	channelRepository "github.com/Hieu3z03/chat-api-golang/modules/channel/repository"
	userRepository "github.com/Hieu3z03/chat-api-golang/modules/user/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChannelService interface {
	Create(ctx context.Context, creatorID uuid.UUID, request dto.CreateChannelRequest) (dto.ChannelResponse, error)
	Get(ctx context.Context, userID, channelID uuid.UUID) (dto.ChannelResponse, error)
	List(ctx context.Context, userID uuid.UUID) ([]dto.ChannelResponse, error)
	EnsureMember(ctx context.Context, channelID, userID uuid.UUID) error
	ListMembersAndMarkRead(
		ctx context.Context,
		channelID, userID uuid.UUID,
		lastReadSequence int64,
	) ([]dto.ChannelMemberReadState, error)
}

type channelService struct {
	channels channelRepository.ChannelRepository
	users    userRepository.UserRepository
}

func NewChannelService(
	channels channelRepository.ChannelRepository,
	users userRepository.UserRepository,
) ChannelService {
	return &channelService{
		channels: channels,
		users:    users,
	}
}

func (service *channelService) Create(
	ctx context.Context,
	creatorID uuid.UUID,
	request dto.CreateChannelRequest,
) (dto.ChannelResponse, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return dto.ChannelResponse{}, dto.ErrInvalidChannelName
	}

	memberIDs := uniqueUserIDs(creatorID, request.MemberIDs)
	users, err := service.users.FindByIDs(ctx, memberIDs)
	if err != nil {
		return dto.ChannelResponse{}, err
	}
	if len(users) != len(memberIDs) {
		return dto.ChannelResponse{}, dto.ErrMembersNotFound
	}

	channel, err := service.channels.Create(ctx, entities.Channel{
		Name:      name,
		CreatedBy: creatorID,
	}, memberIDs)
	if err != nil {
		if errors.Is(err, dto.ErrChannelAlreadyExists) {
			return dto.ChannelResponse{}, dto.ErrChannelAlreadyExists
		}
		return dto.ChannelResponse{}, err
	}

	return toChannelResponse(channel), nil
}

func (service *channelService) Get(
	ctx context.Context,
	userID uuid.UUID,
	channelID uuid.UUID,
) (dto.ChannelResponse, error) {
	if err := service.EnsureMember(ctx, channelID, userID); err != nil {
		return dto.ChannelResponse{}, err
	}

	channel, err := service.channels.FindByID(ctx, channelID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.ChannelResponse{}, dto.ErrChannelNotFound
	}
	if err != nil {
		return dto.ChannelResponse{}, err
	}

	return toChannelResponse(channel), nil
}

func (service *channelService) List(ctx context.Context, userID uuid.UUID) ([]dto.ChannelResponse, error) {
	channels, err := service.channels.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChannelResponse, 0, len(channels))
	for _, channel := range channels {
		responses = append(responses, toChannelResponse(channel))
	}

	return responses, nil
}

func (service *channelService) EnsureMember(ctx context.Context, channelID, userID uuid.UUID) error {
	isMember, err := service.channels.IsMember(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return dto.ErrNotChannelMember
	}

	return nil
}

func (service *channelService) ListMembersAndMarkRead(
	ctx context.Context,
	channelID, userID uuid.UUID,
	lastReadSequence int64,
) ([]dto.ChannelMemberReadState, error) {
	members, err := service.channels.ListMembersAndMarkRead(ctx, channelID, userID, lastReadSequence)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, dto.ErrNotChannelMember
	}

	result := make([]dto.ChannelMemberReadState, 0, len(members))
	for _, member := range members {
		result = append(result, dto.ChannelMemberReadState{
			User: dto.ChannelUserResponse{
				ID:        member.UserID,
				Username:  member.Username,
				Name:      member.Name,
				AvatarURL: member.AvatarURL,
			},
			LastReadSequence: member.LastReadSequence,
			LastReadAt:       member.LastReadAt,
		})
	}

	return result, nil
}

func uniqueUserIDs(creatorID uuid.UUID, requested []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{creatorID: {}}
	result := []uuid.UUID{creatorID}

	for _, userID := range requested {
		if userID == uuid.Nil {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}

		seen[userID] = struct{}{}
		result = append(result, userID)
	}

	return result
}

func toChannelResponse(channel entities.Channel) dto.ChannelResponse {
	members := make([]dto.ChannelUserResponse, 0, len(channel.Members))
	for _, member := range channel.Members {
		members = append(members, dto.ChannelUserResponse{
			ID:        member.User.ID,
			Username:  member.User.Username,
			Name:      member.User.Name,
			AvatarURL: member.User.AvatarURL,
		})
	}

	return dto.ChannelResponse{
		ID:        channel.ID,
		Name:      channel.Name,
		CreatedBy: channel.CreatedBy,
		Members:   members,
		CreatedAt: channel.CreatedAt,
		UpdatedAt: channel.UpdatedAt,
	}
}
