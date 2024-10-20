package migrations

import (
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/user/storage"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&storage.Permission{}, &storage.User{}); err != nil {
		return errors.Wrap(err, "failed to migrate UserDTO and PermissionDTO")
	}
	return nil
}
