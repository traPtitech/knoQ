package repository

import (
	"net/url"
	"strconv"
)

// findMembers グループのメンバーがDBにいるか
// traQIDをもとに探す
func (group *Group) FindMembers() error {
	for i := range group.Members {
		if err := DB.Where("traq_id = ?", group.Members[i].TRAQID).First(&group.Members[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// checkBelongToGroup ユーザーが予約したグループに属しているか調べます
func CheckBelongToGroup(reservationID int, traQID string) (bool, error) {
	rv := new(Reservation)
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
		if traQID == m.TRAQID {
			return true, nil
		}
	}
	return false, nil
}

func FindGroups(values url.Values) ([]Group, error) {
	groups := []Group{}
	cmd := DB.Preload("Members").Preload("CreatedBy")

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

	// SELECT * FROM groups INNER JOIN group_users ON group_users.group_id = groups.id WHERE group_users.user_traq_id = "fuji"
}

func GetGroupIDsBytraQID(traqID string) ([]int, error) {
	groups := []Group{}
	// traqIDが存在するグループを取得
	if err := DB.Raw("SELECT * FROM groups INNER JOIN group_users ON group_users.group_id = groups.id WHERE group_users.user_traq_id =  ?", traqID).Scan(&groups).Error; err != nil {
		return nil, err
	}
	groupsID := make([]int, len(groups))
	for i, g := range groups {
		groupsID[i] = g.ID
	}
	return groupsID, nil
}

// AddRelation add members and created_by by GroupID
func (group *Group) AddRelation(GroupID int) error {
	if err := DB.First(&group, GroupID).Related(&group.Members, "Members").Error; err != nil {
		return err
	}
	if err := group.AddCreatedBy(); err != nil {
		return err
	}
	return nil
}

// AddCreatedBy add CreatedBy
func (group *Group) AddCreatedBy() error {
	if err := DB.Where("traq_id = ?", group.CreatedByRefer).First(&group.CreatedBy).Error; err != nil {
		return err
	}
	return nil
}
