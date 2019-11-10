package repository

import (
	"net/url"
	"strconv"
)

// findMembers グループのメンバーがdbにいるか
// traQIDをもとに探す
func (group *Group) findMembers() error {
	for i := range group.Members {
		if err := db.Where("traq_id = ?", group.Members[i].TRAQID).First(&group.Members[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// checkBelongToGroup ユーザーが予約したグループに属しているか調べます
func checkBelongToGroup(reservationID int, traQID string) (bool, error) {
	rv := new(Reservation)
	g := new(Group)
	rv.ID = reservationID
	if err := db.First(&rv, rv.ID).Error; err != nil {
		return false, err
	}
	g.ID = rv.GroupID
	if err := db.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
		return false, err
	}

	for _, m := range g.Members {
		if traQID == m.TRAQID {
			return true, nil
		}
	}
	return false, nil
}

func findGroups(values url.Values) ([]Group, error) {
	groups := []Group{}
	cmd := db.Preload("Members").Preload("CreatedBy")

	if values.Get("id") != "" {
		id, _ := strconv.Atoi(values.Get("id"))
		cmd = cmd.Where("id = ?", id)
	}

	if values.Get("name") != "" {
		cmd = cmd.Where("name LIKE ?", "%"+values.Get("name")+"%")
	}

	if values.Get("traQID") != "" {
		// traqIDが存在するグループを取得
		groupsID, err := getGroupIDsBytraQID(values.Get("traQID"))
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

func getGroupIDsBytraQID(traqID string) ([]int, error) {
	groups := []Group{}
	// traqIDが存在するグループを取得
	if err := db.Raw("SELECT * FROM groups INNER JOIN group_users ON group_users.group_id = groups.id WHERE group_users.user_traq_id =  ?", traqID).Scan(&groups).Error; err != nil {
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
	if err := db.First(&group, GroupID).Related(&group.Members, "Members").Error; err != nil {
		return err
	}
	if err := group.AddCreatedBy(); err != nil {
		return err
	}
	return nil
}

// AddCreatedBy add CreatedBy
func (group *Group) AddCreatedBy() error {
	if err := db.Where("traq_id = ?", group.CreatedByRefer).First(&group.CreatedBy).Error; err != nil {
		return err
	}
	return nil
}
