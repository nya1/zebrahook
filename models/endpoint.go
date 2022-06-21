package models

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Endpoint struct {
	gorm.Model
	Id              string         `gorm:"primaryKey;not null;<-:create"`
	Url             string         `gorm:"not null"`
	SecretEncrypted string         `gorm:"not null;<-:create;unique_index:uidx_secret_encrypted"`
	EnabledEvents   pq.StringArray `gorm:"type:text[];not null"`
	Metadata        datatypes.JSON `gorm:"type:json;null"`

	Status string `gorm:"not null;default:enabled"`

	// unix timestamp (seconds)
	UpdatedAt int64 `gorm:"autoUpdateTime;not null"`
	CreatedAt int64 `gorm:"autoCreateTime;not null"`
}
