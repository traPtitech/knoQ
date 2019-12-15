package repository

import (
	"errors"
	"net/url"
	"strconv"
)

func (g *Group) Create() error {
	g.ID = 0
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
	err := tx.Set("gorm:association_save_reference", false).Create(&g).Error
	if err != nil {
		tx.Rollback()
		dbErrorLog(err)
		return err
	}
	err = g.verifyMembers()
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Debug().Model(&g).Association("Members").Append(g.Members).Error; err != nil {
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
	if g.ID == 0 {
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

// VerifyUsers グループのメンバーがDBにいるか
// traQIDをもとに探す
// いないものは警告なしに消す
func (g *Group) verifyMembers() error {
	memberSlice := make([]string, 0, len(g.Members))
	for _, v := range g.Members {
		memberSlice = append(memberSlice, v.TRAQID)
	}
	if err := DB.Where(memberSlice).Find(&g.Members).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

// CheckBelongToGroup ユーザーが予約したグループに属しているか調べます
func CheckBelongToGroup(reservationID int, traQID string) (bool, error) {
	rv := new(Event)
	g := new(Group)
	// tmp
	rv.ID = uint64(reservationID)
	if err := DB.First(&rv, rv.ID).Error; err != nil {
		return false, err
	}
	g.ID = rv.GroupID
	if err := DB.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
		return false, err
	}

	for _, m := range g.Members {
		if traQID == m.TRAQID {
			return true, nil
		}
	}
	return false, nil
}

func FindGroups(values url.Values) ([]Group, error) {
	groups := []Group{}
	cmd := DB.Preload("Members")

	if values.Get("id") != "" {
		id, _ := strconv.Atoi(values.Get("id"))
		cmd = cmd.Where("id = ?", id)
	}

	if values.Get("name") != "" {
		cmd = cmd.Where("name LIKE ?", "%"+values.Get("name")+"%")
	}

	if values.Get("traQID") != "" {
		// traqIDが存在するグループを取得
		groupsID, err := GetGroupIDsBytraQID(values.Get("traQID"))
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

func GetGroupIDsBytraQID(traqID string) ([]uint64, error) {
	groups := []Group{}
	// traqIDが存在するグループを取得
	if err := DB.Raw("SELECT * FROM groups INNER JOIN group_users ON group_users.group_id = groups.id WHERE group_users.user_traq_id =  ?", traqID).Scan(&groups).Error; err != nil {
		return nil, err
	}
	groupsID := make([]uint64, len(groups))
	for i, g := range groups {
		groupsID[i] = g.ID
	}
	return groupsID, nil
}

// AddRelation add members and created_by by GroupID
func (group *Group) AddRelation(GroupID uint64) error {
	if err := DB.First(&group, GroupID).Related(&group.Members, "Members").Error; err != nil {
		return err
	}
	return nil
}

// GetCreatedBy get who created it
func (group *Group) GetCreatedBy() (string, error) {
	if err := DB.First(&group).Error; err != nil {
		return "", err
	}
	return group.CreatedBy, nil
}
