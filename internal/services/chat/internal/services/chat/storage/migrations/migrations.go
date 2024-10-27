package migrations

import (
	"github.com/HexArch/go-chat/internal/services/chat/internal/services/chat/storage"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&storage.MessageDTO{}); err != nil {
		return errors.Wrap(err, "failed to migrate MessageDTO")
	}

	if err := db.AutoMigrate(&storage.RoomParticipantDTO{}); err != nil {
		return errors.Wrap(err, "failed to migrate ParticipantDTO")
	}
	return nil
}
