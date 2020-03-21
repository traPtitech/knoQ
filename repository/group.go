package repository

import (
	"errors"
	"fmt"
	"net/url"
	"room/utils"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
)

// WriteGroupParams is used create and update
type WriteGroupParams struct {
	Name        string
	Description string
	ImageID     string
	JoinFreely  bool
	Members     []uuid.UUID
	CreatedBy   uuid.UUID
}

// GroupRepository is implemented by GormRepositoy and API repository.
type GroupRepository interface {
	CreateGroup(groupParams WriteGroupParams) (*Group, error)
	UpdateGroup(groupID uuid.UUID, groupParams WriteGroupParams) (*Group, error)
	// AddUserToGroup add a user to that group if that group is open.
	AddUserToGroup(groupID uuid.UUID, userID uuid.UUID) error
	DeleteGroup(groupID uuid.UUID) error
	// DeleteUserInGroup delete a user in that group if that group is open.
	DeleteUserInGroup(groupID uuid.UUID, userID uuid.UUID) error
	GetGroup(groupID uuid.UUID) (*Group, error)
	GetAllGroups() ([]*Group, error)
	GetUserBelongingGroupIDs(userID uuid.UUID) ([]uuid.UUID, error)
}

// GormRepository implements GroupRepository

// CreateGroup create Group in DB
func (repo *GormRepository) CreateGroup(groupParams WriteGroupParams) (*Group, error) {
	group := new(Group)
	err := copier.Copy(&group, groupParams)
	if err != nil {
		return nil, err
	}
	group.Members, err = verifyuserIDs(repo.DB, groupParams.Members)
	if err != nil {
		return nil, err
	}

	if err = repo.DB.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

// UpdateGroup update group in DB.
func (repo *GormRepository) UpdateGroup(groupID uuid.UUID, groupParams WriteGroupParams) (*Group, error) {
	if groupID == uuid.Nil {
		return nil, ErrNilID
	}
	group := new(Group)
	err := copier.Copy(&group, groupParams)
	if err != nil {
		return nil, err
	}
	group.Members, err = verifyuserIDs(repo.DB, groupParams.Members)
	if err != nil {
		return nil, err
	}

	group.ID = groupID

	if err = repo.DB.Save(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

// AddUserToGroup add user to group
func (repo *GormRepository) AddUserToGroup(groupID uuid.UUID, userID uuid.UUID) error {
	if userID == uuid.Nil || groupID == uuid.Nil {
		return ErrNilID
	}
	return repo.DB.Transaction(func(tx *gorm.DB) error {
		group := new(Group)
		if err := tx.Preload("Members").Where("id = ?", groupID).First(&group).Error; err != nil {
			return err
		}
		member, err := verifyuserID(tx, userID)
		if err != nil {
			return err
		}
		if group.IsMember(member) {
			return ErrAlreadyExists
		}
		if err := tx.Model(&Group{ID: groupID}).Association("Members").Append(member).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteGroup soft deletes group column
// and does not delete group members
func (repo *GormRepository) DeleteGroup(groupID uuid.UUID) error {
	if groupID == uuid.Nil {
		return ErrNilID
	}
	result := repo.DB.Where("id = ?", groupID).Delete(&Group{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *GormRepository) DeleteUserInGroup(groupID uuid.UUID, userID uuid.UUID) error {
	if userID == uuid.Nil || groupID == uuid.Nil {
		return ErrNilID
	}

	return repo.DB.Transaction(func(tx *gorm.DB) error {
		group := new(Group)
		if err := tx.Preload("Members").Where("id = ?", groupID).First(&group).Error; err != nil {
			return err
		}
		member, err := verifyuserID(tx, userID)
		if err != nil {
			return err
		}
		if !group.IsMember(member) {
			return ErrNotFound
		}
		if err := tx.Model(&Group{ID: groupID}).Association("Members").Delete(member).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetGroup gets group with members
func (repo *GormRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	if groupID == uuid.Nil {
		return nil, ErrNilID
	}

	cmd := repo.DB.Preload("Members")
	group := new(Group)
	if err := cmd.Where("id = ?", groupID).Take(&group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

// GetAllGroups gets all groups with members
func (repo *GormRepository) GetAllGroups() ([]*Group, error) {
	groups := make([]*Group, 0)
	cmd := repo.DB.Preload("Members")

	if err := cmd.Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// GetUserBelongingGroupIDs gets group IDs
func (repo *GormRepository) GetUserBelongingGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	groupIDs := make([]uuid.UUID, 0)

	// userIDが存在するグループを取得
	rows, err := repo.DB.Table("group_users").Select("group_id").Where("user_id = ?", userID).Rows()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var groupID uuid.UUID
		rows.Scan(&groupID)
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, nil
}

// TraQRepository

// CreateGroup always return error
func (repo *TraQRepository) CreateGroup(groupParams WriteGroupParams) (*Group, error) {
	return nil, ErrForbidden
}

// UpdateGroup always return error
func (repo *TraQRepository) UpdateGroup(groupID uuid.UUID, groupParams WriteGroupParams) (*Group, error) {
	return nil, ErrForbidden
}

// AddUserToGroup always return error
func (repo *TraQRepository) AddUserToGroup(groupID uuid.UUID, userID uuid.UUID) error {
	return ErrForbidden
}

// DeleteGroup always return error
func (repo *TraQRepository) DeleteGroup(groupID uuid.UUID) error {
	return ErrForbidden
}

// DeleteUserInGroup always return error
func (repo *TraQRepository) DeleteUserInGroup(groupID uuid.UUID, userID uuid.UUID) error {
	return ErrForbidden
}

func (repo *TraQRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	if groupID == uuid.Nil {
		return nil, ErrNilID
	}
	data, err := utils.APIGetRequest(repo.Token, fmt.Sprintf("/groups/%s", groupID))
	if err != nil {
		return nil, err
	}
	group := new(Group)
	err = traQjson.Unmarshal(data, &group)
	return group, err
}

func (repo *TraQRepository) GetAllGroups() ([]*Group, error) {
	data, err := utils.APIGetRequest(repo.Token, "/groups")
	if err != nil {
		return nil, err
	}
	groups := make([]*Group, 0)
	err = traQjson.Unmarshal(data, &groups)
	return groups, err
}

func (repo *TraQRepository) GetUserBelongingGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	data, err := utils.APIGetRequest(repo.Token, fmt.Sprintf("/users/%s/groups", userID))
	if err != nil {
		return nil, err
	}
	groupIDs := make([]uuid.UUID, 0)
	err = traQjson.Unmarshal(data, &groupIDs)
	return groupIDs, err
}

func verifyuserID(db *gorm.DB, userID uuid.UUID) (*User, error) {
	member := new(User)
	if err := db.Where("id = ?", userID).Take(&member).Error; err != nil {
		return nil, err
	}
	return member, nil
}

func verifyuserIDs(db *gorm.DB, userIDs []uuid.UUID) ([]User, error) {
	members := []User{}
	if err := db.Where("id IN (?)", userIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (g *Group) IsMember(user *User) bool {
	for _, member := range g.Members {
		if member.ID == user.ID {
			return true
		}
	}
	return false
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
