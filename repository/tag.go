package repository

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// MatchTags なければ $attr = true で作成、あっても $attr = true に更新
func MatchTags(tags []Tag, attr string) error {
	for i := range tags {
		tag := &tags[i]
		if err := MatchTag(tag, attr); err != nil {
			return err
		}
	}
	return nil
}

func MatchTag(tag *Tag, attr string) error {
	createTag, attrFlag, updateAttr, err := judgeTagAttr(*tag, attr)
	if err != nil {
		return err
	}

	if err := DB.Where(Tag{Name: tag.Name}).First(tag).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			dbErrorLog(err)
			return err
		}
		*tag = createTag
		err := DB.Create(&tag).Error
		if err != nil {
			dbErrorLog(err)
			return err
		}
	}
	if attrFlag {
		err := DB.Debug().Model(&tag).Update(updateAttr).Error
		if err != nil {
			dbErrorLog(err)
			return err
		}
	}
	return nil
}

func judgeTagAttr(tag Tag, attr string) (createTag Tag, attrFlag bool, updateAttr map[string]interface{}, err error) {
	switch attr {
	case "event":
		createTag = Tag{Name: tag.Name, ForEvent: true}
		attrFlag = !tag.ForEvent
		updateAttr = map[string]interface{}{"for_event": true}
	case "group":
		createTag = Tag{Name: tag.Name, ForGroup: true}
		attrFlag = !tag.ForGroup
		updateAttr = map[string]interface{}{"for_group": true}
	case "room":
		createTag = Tag{Name: tag.Name, ForRoom: true}
		attrFlag = !tag.ForGroup
		updateAttr = map[string]interface{}{"for_room": true}
	default:
		err = errors.New("this attr does not exist")
	}
	return
}
