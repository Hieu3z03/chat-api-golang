package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	FirstName string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName  string     `gorm:"type:varchar(100);not null" json:"last_name"`
	Username  string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	AvatarID  *uuid.UUID `gorm:"type:uuid" json:"avatar_id,omitempty"`

	Timestamp
}

func (user *User) BeforeCreate(_ *gorm.DB) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	return nil
}
