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

func (g *Group) Read() error {
	cmd := DB.Preload("Members")
	if err := cmd.First(&g).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
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
