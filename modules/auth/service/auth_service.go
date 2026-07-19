package service

import (
	"context"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/auth/dto"
	authRepo "github.com/Hieu3z03/chat-api-golang/modules/auth/repository"
	userDto "github.com/Hieu3z03/chat-api-golang/modules/user/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/user/repository"
	"github.com/Hieu3z03/chat-api-golang/pkg/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	Login(ctx context.Context, req userDto.UserLoginRequest) (dto.TokenResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error)
	Logout(ctx context.Context, userId string) error
}

type authService struct {
	userRepository         repository.UserRepository
	refreshTokenRepository authRepo.RefreshTokenRepository
	jwtService             JWTService
	db                     *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo authRepo.RefreshTokenRepository,
	jwtService JWTService,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository:         userRepo,
		refreshTokenRepository: refreshTokenRepo,
		jwtService:             jwtService,
		db:                     db,
	}
}

func (s *authService) Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return userDto.UserResponse{}, err
	}

	if isExist {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
	}

	user := entities.User{
		ID:         uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		TelpNumber: req.TelpNumber,
		Password:   req.Password,
		Role:       "user",
	}

	createdUser, err := s.userRepository.Register(ctx, s.db, user)
	if err != nil {
		return userDto.UserResponse{}, err
	}

	return userDto.UserResponse{
		ID:         createdUser.ID.String(),
		Name:       createdUser.Name,
		Email:      createdUser.Email,
		TelpNumber: createdUser.TelpNumber,
		Role:       createdUser.Role,
	}, nil
}

func (s *authService) Login(ctx context.Context, req userDto.UserLoginRequest) (dto.TokenResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return dto.TokenResponse{}, userDto.ErrEmailNotFound
	}

	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return dto.TokenResponse{}, dto.ErrInvalidCredentials
	}

	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)
	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	refreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, refreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         user.Role,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error) {
	refreshToken, err := s.refreshTokenRepository.FindByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return dto.TokenResponse{}, dto.ErrRefreshTokenNotFound
	}

	accessToken := s.jwtService.GenerateAccessToken(refreshToken.UserID.String(), refreshToken.User.Role)
	newRefreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	err = s.refreshTokenRepository.DeleteByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	newRefreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    refreshToken.UserID,
		Token:     newRefreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, newRefreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		Role:         refreshToken.User.Role,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userId string) error {
	return s.refreshTokenRepository.DeleteByUserID(ctx, s.db, userId)
}
