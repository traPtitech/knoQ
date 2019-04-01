package main

import (
	"fmt"
	"strconv"
)

// getUser ユーザー情報を取得します
func getUser(id string) (*User, error) {
	user := User{}

	// DBに登録されていない場合(初めてアクセスした場合)はDBにレコードを作成する
	if err := db.FirstOrCreate(&user, &User{TRAQID: id}).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// changeUserToAdmin ユーザーの管理者権限の有無を変更します
func changeUserToAdmin(id string, isAdmin bool) error {
	// ユーザー取得
	user, err := getUser(id)
	if err != nil {
		return err
	}

	// 変更
	if err := db.Model(user).Update("admin", isAdmin).Error; err != nil {
		return err
	}
	return nil
}

// checkMembers グループのメンバーがdbにいるか
func checkMembers(group *Group) error {
	for _, u := range group.Members {
		if err := db.Where("traq_id = ?", u.TRAQID).First(&u).Error; err != nil {
			return err
		}
	}
	return nil
}

func checkGroup(groupID int) error {
	g := new(Group)
	g.ID = groupID
	if err := db.First(&g, g.ID).Error; err != nil {
		return err
	}
	return nil
}

func checkRoom(roomID int) error {
	r := new(Room)
	r.ID = roomID
	if err := db.First(&r, r.ID).Error; err != nil {
		return err
	}
	return nil
}

func getRooms(begin, end string) ([]Room, error) {
	rooms := []Room{}
	cmd := db
	if begin != "" {
		cmd = cmd.Where("date >= ?", begin)
	}
	if end != "" {
		cmd = cmd.Where("date <= ?", end)
	}

	if err := cmd.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func getUserBelongGroups(traqID string) ([]Group, error) {
	groups := []Group{}
	if err := db.Find(&groups).Error; err != nil {
		return nil, err
	}
	resGroups := []Group{}
	for _, g := range groups {
		if err := db.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
			return nil, err
		}
		for _, user := range g.Members {
			if user.TRAQID == traqID || traqID == "" {
				resGroups = append(resGroups, g)
				break
			}
		}
	}
	return resGroups, nil
}

func findRvs(traqID, groupID, begin, end string) ([]Reservation, error) {
	reservations := []Reservation{}
	cmd := db
	groupIDs := []int{}

	groups, err := getUserBelongGroups(traqID)
	if err != nil {
		return nil, err
	}

	if groupID != "" {
		groupid, _ := strconv.Atoi(groupID)
		for _, g := range groups {
			if groupid == g.ID {
				groupIDs = append(groupIDs, g.ID)
			}
		}
	} else {
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}
	}

	cmd = cmd.Where("group_id in (?)", groupIDs)

	if begin != "" {
		cmd = cmd.Where("date > ?", begin)
		fmt.Println(begin)
	}
	if end != "" {
		cmd = cmd.Where("date <= ?", end)
		fmt.Println(end)
	}

	fmt.Println(cmd)
	if err := cmd.Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}
