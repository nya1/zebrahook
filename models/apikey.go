package models

import "gorm.io/gorm"

type ApiKey struct {
	gorm.Model
	Id          uint   `gorm:"primaryKey;not null;<-:create"`
	Description string `gorm:"not null"`
	Hash        string `gorm:"not null;<-:create;unique_index:uidx_apikey_hash"`

	Status string `gorm:"not null;default:enabled"`

	// unix timestamp (seconds)
	UpdatedAt int64 `gorm:"autoUpdateTime;not null"`
	CreatedAt int64 `gorm:"autoCreateTime;not null"`
}
