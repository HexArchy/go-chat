package migrations

import (
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&storage.Token{}); err != nil {
		return errors.Wrap(err, "failed to migrate TokenDTO")
	}
	return nil
}
