package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Name      string    `gorm:"type:varchar(200);not null" json:"name"`
	AvatarURL *string   `gorm:"type:varchar(2048)" json:"avatar_url,omitempty"`
}

func (user *User) BeforeCreate(_ *gorm.DB) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	return nil
}
