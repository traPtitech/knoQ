package model

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	MARIADB_HOSTNAME = os.Getenv("MARIADB_HOSTNAME")
	MARIADB_DATABASE = os.Getenv("MARIADB_DATABASE")
	MARIADB_USERNAME = os.Getenv("MARIADB_USERNAME")
	MARIADB_PASSWORD = os.Getenv("MARIADB_PASSWORD")

	db *gorm.DB
)

// SetupDatabase set up db and crate tables
func SetupDatabase() (*gorm.DB, error) {
	var err error
	//tmp
	if MARIADB_HOSTNAME == "" {
		MARIADB_HOSTNAME = ""
	}
	if MARIADB_DATABASE == "" {
		MARIADB_DATABASE = "room"
	}
	if MARIADB_USERNAME == "" {
		MARIADB_USERNAME = "root"
	}

	if MARIADB_PASSWORD == "" {
		MARIADB_PASSWORD = "password"
	}

	// データベース接続
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", MARIADB_USERNAME, MARIADB_PASSWORD, MARIADB_HOSTNAME, MARIADB_DATABASE))
	if err != nil {
		return db, err
	}
	if err := initDB(); err != nil {
		return db, err
	}
	return db, nil
}

// initDB データベースのスキーマを更新
func initDB() error {
	// テーブルが無ければ作成
	if err := db.AutoMigrate(tables...).Error; err != nil {
		return err
	}
	return nil
}

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

	if err := cmd.Order("date asc").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
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

func findRvs(values url.Values) ([]Reservation, error) {
	reservations := []Reservation{}
	cmd := db.Preload("Group").Preload("Group.Members").Preload("Group.CreatedBy").Preload("Room").Preload("CreatedBy")

	if values.Get("id") != "" {
		id, _ := strconv.Atoi(values.Get("id"))
		cmd = cmd.Where("id = ?", id)
	}

	if values.Get("name") != "" {
		cmd = cmd.Where("name LIKE ?", "%"+values.Get("name")+"%")
	}

	if values.Get("traQID") != "" {
		groupsID, err := getGroupIDsBytraQID(values.Get("traQID"))
		if err != nil {
			return nil, err
		}
		cmd = cmd.Where("group_id in (?)", groupsID)
	}

	if values.Get("groupid") != "" {
		groupid, _ := strconv.Atoi(values.Get("groupid"))
		cmd = cmd.Where("group_id = ?", groupid)
	}

	if values.Get("roomid") != "" {
		roomid, _ := strconv.Atoi(values.Get("roomid"))
		cmd = cmd.Where("room_id = ?", roomid)
	}

	if values.Get("date_begin") != "" {
		cmd = cmd.Where("date >= ?", values.Get("date_begin"))
	}
	if values.Get("date_end") != "" {
		cmd = cmd.Where("date <= ?", values.Get("date_end"))
	}

	if err := cmd.Order("date asc").Find(&reservations).Error; err != nil {
		return nil, err
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
