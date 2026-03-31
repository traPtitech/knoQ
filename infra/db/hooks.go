package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)



// 妥当
func (et *EventTag) BeforeDelete(tx *gorm.DB) (err error) {
	// タグのIDが空で名前が提供されている場合は、
	// 名前に応じたタグを削除する
	if et.TagID == uuid.Nil && et.Tag.Name != "" {
		tag := Tag{
			Name: et.Tag.Name,
		}
		err = tx.Where(&tag).Take(&tag).Error
		if err != nil {
			return err
		}

		et.TagID = tag.ID
	}
	return nil
}