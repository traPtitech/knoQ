package repository

import (
	"github.com/jinzhu/gorm"
)

// MatchEventTag なければ for_event = true で作成、あっても for_event = true に更新
func MatchEventTags(tags []Tag) error {
	for i := range tags {
		tag := &tags[i]
		if err := MatchEventTag(tag); err != nil {
			return err
		}
	}
	return nil
}

func MatchEventTag(tag *Tag) error {
	if err := DB.Where(Tag{Name: tag.Name}).First(&tag).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			dbErrorLog(err)
			return err
		}
		err := DB.Create(&Tag{Name: tag.Name, ForEvent: true}).Error
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
	return nil
}
