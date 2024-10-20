package migrations

import (
	"github.com/HexArch/go-chat/internal/services/website/internal/services/rooms/storage"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&storage.Room{}); err != nil {
		return errors.Wrap(err, "failed to migrate rooms table")
	}
	return nil
}
