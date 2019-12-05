package repository

import (
	"github.com/jinzhu/gorm"
)

// MatchEventTag なければ for_event = true で作成、あっても for_event = true に更新
func MatchEventTag(tags []Tag) error {
	for i, t := range tags {
		tag := &tags[i]
		if err := DB.Where(Tag{Name: t.Name}).First(&tag).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				dbErrorLog(err)
				return err
			}
			err := DB.Create(&Tag{Name: t.Name, ForEvent: true}).Error
			if err != nil {
				dbErrorLog(err)
				return err
			}

		}
		if !tag.ForEvent {
			err := DB.Model(&tag).Update("for_event", true).Error
			if err != nil {
				dbErrorLog(err)
				return err
			}
		}
	}
	return nil
}
