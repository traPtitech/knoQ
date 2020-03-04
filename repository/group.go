package repository

import (
	"errors"
	"net/url"

	"github.com/gofrs/uuid"
)

// GormRepostory and API repository implement GroupRepository.
type GroupRepository interface {
}

func (g *Group) Create() error {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		dbErrorLog(err)
		return err
	}
	err := g.verifyMembers()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Set("gorm:save_associations", false).Create(&g).Error
	if err != nil {
		tx.Rollback()
		dbErrorLog(err)
		return err
	}

	return tx.Commit().Error
}

func (g *Group) Read() error {
	cmd := DB.Preload("Members")
	if err := cmd.First(&g).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

func (g *Group) Update() error {
	nowGroup := new(Group)
	nowGroup.ID = g.ID
	if err := nowGroup.Read(); err != nil {
		return err
	}
	g.CreatedAt = nowGroup.CreatedAt
	g.CreatedBy = nowGroup.CreatedBy

	if err := g.verifyMembers(); err != nil {
		return err
	}

	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		dbErrorLog(err)
		return err
	}

	if err := tx.Debug().Save(&g).Error; err != nil {
		tx.Rollback()
		dbErrorLog(err)
		return err
	}
	if err := tx.Debug().Model(&g).Association("Members").Replace(g.Members).Error; err != nil {
		tx.Rollback()
		dbErrorLog(err)
		return err
	}

	return tx.Commit().Error
}

func (g *Group) Delete() error {
	if g.ID == uuid.Nil {
		err := errors.New("ID=0. You want to Delete All ?")
		dbErrorLog(err)
		return err
	}
	if err := g.Read(); err != nil {
		return err
	}
	if err := DB.Debug().Delete(&g).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

// BeforeCreate is gorm hook
func (g *Group) BeforeCreate() (err error) {
	g.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// VerifyUsers グループのメンバーがDBにいるか
// traQIDをもとに探す
// いないものは警告なしに消す
func (g *Group) verifyMembers() error {
	memberSlice := make([]uuid.UUID, 0, len(g.Members))
	for _, v := range g.Members {
		memberSlice = append(memberSlice, v.ID)
	}
	g.Members = nil
	if err := DB.Debug().Where("id IN (?)", memberSlice).Find(&g.Members).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

// CheckBelongToGroup ユーザーが予約したグループに属しているか調べます
func CheckBelongToGroup(reservationID uuid.UUID, traQID uuid.UUID) (bool, error) {
	rv := new(Event)
	g := new(Group)
	rv.ID = reservationID
	if err := DB.First(&rv, rv.ID).Error; err != nil {
		return false, err
	}
	g.ID = rv.GroupID
	if err := DB.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
		return false, err
	}

	for _, m := range g.Members {
		if traQID == m.ID {
			return true, nil
		}
	}
	return false, nil
}

func FindGroups(values url.Values) ([]Group, error) {
	groups := []Group{}
	cmd := DB.Preload("Members")

	if values.Get("userID") != "" {
		// userIDが存在するグループを取得
		groupsID, err := GetGroupIDsBytraQID(values.Get("userID"))
		if err != nil {
			return nil, err
		}
		cmd = cmd.Where(groupsID)
	}

	if err := cmd.Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

func GetGroupIDsBytraQID(traqID string) ([]uuid.UUID, error) {
	groups := []Group{}
	// traqIDが存在するグループを取得
	if err := DB.Raw("SELECT * FROM groups INNER JOIN group_users ON group_users.group_id = groups.id WHERE group_users.user_traq_id =  ?", traqID).Scan(&groups).Error; err != nil {
		return nil, err
	}
	groupsID := make([]uuid.UUID, len(groups))
	for i, g := range groups {
		groupsID[i] = g.ID
	}
	return groupsID, nil
}

// AddRelation add members and created_by by GroupID
func (group *Group) AddRelation(GroupID uuid.UUID) error {
	if err := DB.First(&group, GroupID).Related(&group.Members, "Members").Error; err != nil {
		return err
	}
	return nil
}

// GetCreatedBy get who created it
func (group *Group) GetCreatedBy() (uuid.UUID, error) {
	if err := DB.First(&group).Error; err != nil {
		return uuid.Nil, err
	}
	return group.CreatedBy, nil
}

// AddTag add tag
/*
func (g *Group) AddTag(tagName string, locked bool) error {
	tag := new(Tag)
	tag.Name = tagName

	if err := MatchTag(tag, "group"); err != nil {
		return err
	}
	if err := DB.Create(&GroupTag{GroupID: g.ID, TagID: tag.ID, Locked: locked}).Error; err != nil {
		return err
	}
	return nil
}

// DeleteTag delete unlocked tag.
func (g *Group) DeleteTag(tagID uuid.UUID) error {
	groupTag := new(GroupTag)
	groupTag.TagID = tagID
	groupTag.GroupID = g.ID
	if err := DB.Debug().First(&groupTag).Error; err != nil {
		return err
	}
	if groupTag.Locked {
		return errors.New("this tag is locked")
	}
	if err := DB.Debug().Where("locked = ?", false).Delete(&GroupTag{GroupID: g.ID, TagID: groupTag.TagID}).Error; err != nil {
		return err
	}
	return nil
}
*/

func (g *Group) AddMember(userID uuid.UUID) error {
	user := new(User)
	user.ID = userID
	if err := DB.First(&user).Error; err != nil {
		dbErrorLog(err)
		return err
	}

	if err := DB.Model(&g).Association("Members").Append(user).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

func (g *Group) DeleteMember(userID uuid.UUID) error {
	user := new(User)
	user.ID = userID
	if err := DB.First(&user).Error; err != nil {
		dbErrorLog(err)
		return err
	}

	if err := DB.Model(&g).Association("Members").Delete(user).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}
