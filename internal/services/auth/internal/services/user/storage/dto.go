package storage

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id"`
	Email       string        `gorm:"uniqueIndex;column:email"`
	Password    string        `gorm:"column:password"`
	Username    string        `gorm:"column:username"`
	Phone       string        `gorm:"column:phone"`
	Age         int           `gorm:"column:age"`
	Bio         string        `gorm:"column:bio"`
	Permissions []*Permission `gorm:"many2many:user_permissions;"`
	CreatedAt   time.Time     `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime;column:updated_at"`
}

type Permission struct {
	ID   uint   `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"uniqueIndex;column:name"`
}
