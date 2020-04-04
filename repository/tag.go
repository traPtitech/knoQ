package repository

import (
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
)

type TagRepository interface {
	CreateOrGetTag(name string) (*Tag, error)
	GetTagByName(name string) (*Tag, error)
	// UpdateTag(tagID uuid.UUID, name string) (*Tag, error)
	// DeleteTag(tagID uuid.UUID) error
	GetTag(tagID uuid.UUID) (*Tag, error)
	GetAllTags() ([]*Tag, error)
}

func (repo *GormRepository) CreateOrGetTag(name string) (*Tag, error) {
	tag := new(Tag)

	if err := repo.DB.Where("name = ?", name).Take(&tag).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		tag.Name = name
		err = repo.DB.Create(&tag).Error
		if err != nil {
			return nil, err
		}
	}
	return tag, nil
}

func (repo *GormRepository) GetTagByName(name string) (*Tag, error) {
	tag := new(Tag)
	err := repo.DB.Where("name = ?", name).Take(&tag).Error
	return tag, err
}

func (repo *GormRepository) GetTag(tagID uuid.UUID) (*Tag, error) {
	tag := new(Tag)
	tag.ID = tagID
	err := repo.DB.Take(&tag).Error
	return tag, err
}

func (repo *GormRepository) GetAllTags() ([]*Tag, error) {
	tags := make([]*Tag, 0)
	err := repo.DB.Find(&tags).Error
	return tags, err
}

func (t *Tag) Create() error {
	t.Name = strings.ToLower(t.Name)

	if err := DB.Create(&t).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

// FindTags return all tags
func FindTags() ([]Tag, error) {
	tags := []Tag{}
	if err := DB.First(&tags).Error; err != nil {
		dbErrorLog(err)
		return nil, err
	}
	return tags, nil
}

// BeforeCreate is gorm hook
func (t *Tag) BeforeCreate() (err error) {
	t.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

/*
func MatchTags(tags []Tag, attr string) error {
	for i := range tags {
		tag := &tags[i]
		if err := MatchTag(tag, attr); err != nil {
			return err
		}
	}
	return nil
}
*/

// MatchTag なければ $attr = true で作成、あっても $attr = true に更新
/*
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
*/

/*
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
*/
