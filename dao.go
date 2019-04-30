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

func findRoomsByTime(begin, end string) ([]Room, error) {
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

func findGroupsBelong(traqID string) ([]Group, error) {
	groups := []Group{}
	// groupsを全取得
	if err := db.Find(&groups).Error; err != nil {
		return nil, err
	}
	resGroups := []Group{}
	for _, g := range groups {
		// membersを紐付ける
		if err := db.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
			return nil, err
		}
		for _, user := range g.Members {
			// requestに合ったtraQIDのみを追加する
			if user.TRAQID == traqID || traqID == "" {
				if err := db.Where("traq_id = ?", g.CreatedByRefer).First(&g.CreatedBy).Error; err != nil {
					return nil, err
				}
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

	groups, err := findGroupsBelong(traqID)
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
	// relationの追加
	for i := range reservations {
		group := &reservations[i].Group
		room := &reservations[i].Room
		// group
		group.AddRelation(reservations[i].GroupID)

		// room
		if err := db.First(&room, reservations[i].RoomID).Error; err != nil {
			return nil, err
		}

		// createdBy
		if err := db.Where("traq_id = ?", reservations[i].CreatedByRefer).First(&reservations[i].CreatedBy).Error; err != nil {
			return nil, err
		}
	}

	return reservations, nil
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

// AddCreatedBy add CreatedBy
func (reservation *Reservation) AddCreatedBy() error {
	if err := db.Where("traq_id = ?", reservation.CreatedByRefer).First(&reservation.CreatedBy).Error; err != nil {
		return err
	}
	return nil
}

// AddRelation add room by RoomID
func (room *Room) AddRelation(roomID int) error {
	room.ID = roomID
	if err := db.First(&room, room.ID).Error; err != nil {
		return err
	}
	return nil
}
