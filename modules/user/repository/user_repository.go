package repository

import (
	"context"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Upsert(ctx context.Context, user entities.User) (entities.User, error)
	FindByID(ctx context.Context, userID uuid.UUID) (entities.User, error)
	FindByUsername(ctx context.Context, username string) (entities.User, error)
	FindByIDs(ctx context.Context, userIDs []uuid.UUID) ([]entities.User, error)
	Search(ctx context.Context, search string, limit int) ([]entities.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (repository *userRepository) Upsert(ctx context.Context, user entities.User) (entities.User, error) {
	defer appLogger.Measure(ctx, "UserRepository.Upsert")()

	err := repository.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"username",
			"name",
			"avatar_url",
		}),
	}).Create(&user).Error
	if err != nil {
		return entities.User{}, err
	}

	return repository.FindByID(ctx, user.ID)
}

func (repository *userRepository) FindByID(ctx context.Context, userID uuid.UUID) (entities.User, error) {
	defer appLogger.Measure(ctx, "UserRepository.FindByID")()

	var user entities.User
	if err := repository.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (repository *userRepository) FindByUsername(ctx context.Context, username string) (entities.User, error) {
	defer appLogger.Measure(ctx, "UserRepository.FindByUsername")()

	var user entities.User
	if err := repository.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (repository *userRepository) FindByIDs(ctx context.Context, userIDs []uuid.UUID) ([]entities.User, error) {
	defer appLogger.Measure(ctx, "UserRepository.FindByIDs")()

	var users []entities.User
	if len(userIDs) == 0 {
		return users, nil
	}

	if err := repository.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (repository *userRepository) Search(ctx context.Context, search string, limit int) ([]entities.User, error) {
	defer appLogger.Measure(ctx, "UserRepository.Search")()

	query := repository.db.WithContext(ctx).Order("username ASC").Limit(limit)
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR username ILIKE ?",
			pattern,
			pattern,
		)
	}

	var users []entities.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
