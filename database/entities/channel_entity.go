package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Channel struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string          `gorm:"type:varchar(120);not null" json:"name"`
	CreatedBy uuid.UUID       `gorm:"type:uuid;not null;index" json:"created_by"`
	Creator   User            `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"creator,omitempty"`
	Members   []ChannelMember `gorm:"foreignKey:ChannelID" json:"members,omitempty"`

	Timestamp
}

func (channel *Channel) BeforeCreate(_ *gorm.DB) error {
	if channel.ID == uuid.Nil {
		channel.ID = uuid.New()
	}

	return nil
}

type ChannelMember struct {
	ChannelID uuid.UUID `gorm:"type:uuid;primaryKey" json:"channel_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	JoinedAt  time.Time `gorm:"type:timestamp with time zone;not null;autoCreateTime" json:"joined_at"`
	Channel   Channel   `gorm:"foreignKey:ChannelID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
}
