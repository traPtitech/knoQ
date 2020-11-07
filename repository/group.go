package repository

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	traQrouterV3 "github.com/traPtitech/traQ/router/v3"
)

var traPGroupID = uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))

// WriteGroupParams is used create and update
type WriteGroupParams struct {
	Name        string
	Description string
	JoinFreely  bool
	Members     []uuid.UUID
	Admins      []uuid.UUID
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
	if err = repo.DB.Create(group).Error; err != nil {
		return nil, err
	}
	group.Members = formatGroupMembers(group.ID, groupParams.Members)
	group.Admins = formatGroupAdmins(group.ID, groupParams.Admins)
	if err != nil {
		return nil, err
	}
	for _, m := range group.Members {
		if err = repo.DB.Create(m).Error; err != nil {
			return nil, err
		}
	}
	for _, a := range group.Admins {
		if err = repo.DB.Create(a).Error; err != nil {
			return nil, err
		}
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
	group.Members = formatGroupMembers(groupID, groupParams.Members)
	group.Admins = formatGroupAdmins(group.ID, groupParams.Admins)
	if err != nil {
		return nil, err
	}

	group.ID = groupID

	if err = repo.DB.Save(group).Error; err != nil {
		return nil, err
	}
	for _, m := range group.Members {
		if err = repo.DB.Save(m).Error; err != nil {
			return nil, err
		}
	}
	for _, a := range group.Admins {
		if err = repo.DB.Save(a).Error; err != nil {
			return nil, err
		}
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
		if !group.JoinFreely {
			return ErrForbidden
		}
		member := &GroupUsers{GroupID: groupID, UserID: userID}
		if group.IsMember(member) {
			return ErrAlreadyExists
		}
		if err := tx.Create(member).Error; err != nil {
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
		if !group.JoinFreely {
			return ErrForbidden
		}
		member := &GroupUsers{UserID: userID}
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

	cmd := repo.DB.Preload("Members").Preload("Admins")
	group := new(Group)
	if err := cmd.Where("id = ?", groupID).Take(&group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

// GetAllGroups gets all groups with members
func (repo *GormRepository) GetAllGroups() ([]*Group, error) {
	groups := make([]*Group, 0)
	cmd := repo.DB.Preload("Members").Preload("Admins")

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
	if repo.Version != TraQv3 {
		return nil, ErrForbidden
	}
	reqGroup := &traQrouterV3.PostUserGroupRequest{
		Name:        groupParams.Name,
		Description: groupParams.Description,
		Type:        "room",
	}
	body, _ := json.Marshal(reqGroup)
	resBody, err := repo.postRequest("/groups", body)
	if err != nil {
		return nil, err
	}
	traQgroup := new(traQrouterV3.UserGroup)
	err = json.Unmarshal(resBody, &traQgroup)
	if err != nil {
		return nil, err
	}
	group := formatV3Group(traQgroup)
	for _, userID := range groupParams.Members {
		reqMember := &traQrouterV3.PostUserGroupMemberRequest{
			ID: userID,
		}
		body, _ := json.Marshal(reqMember)
		_, err := repo.postRequest(fmt.Sprintf("/groups/%s/members", group.ID), body)
		if err != nil {
			return nil, err
		}
		group.Members = append(group.Members, GroupUsers{UserID: userID})
	}

	return group, nil
}

// UpdateGroup always return error
func (repo *TraQRepository) UpdateGroup(groupID uuid.UUID, groupParams WriteGroupParams) (*Group, error) {
	return nil, ErrForbidden
}

// AddUserToGroup always return error
func (repo *TraQRepository) AddUserToGroup(groupID uuid.UUID, userID uuid.UUID) error {
	if repo.Version != TraQv3 {
		return ErrForbidden
	}

	reqMember := &traQrouterV3.PostUserGroupMemberRequest{
		ID: userID,
	}
	body, _ := json.Marshal(reqMember)
	_, err := repo.postRequest(fmt.Sprintf("/groups/%s/members", groupID), body)
	return err
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
	if repo.Version != TraQv3 {
		return nil, ErrForbidden
	}
	if groupID == uuid.Nil {
		return nil, ErrNilID
	}
	data, err := repo.getRequest(fmt.Sprintf("/groups/%s", groupID))
	if err != nil {
		return nil, err
	}
	traQgroup := new(traQrouterV3.UserGroup)
	err = json.Unmarshal(data, &traQgroup)
	return formatV3Group(traQgroup), err
}

func (repo *TraQRepository) GetAllGroups() ([]*Group, error) {
	if repo.Version != TraQv3 {
		return nil, ErrForbidden
	}

	data, err := repo.getRequest("/groups")
	if err != nil {
		return nil, err
	}
	traQgroups := make([]*traQrouterV3.UserGroup, 0)
	err = traQjson.Unmarshal(data, &traQgroups)
	groups := make([]*Group, len(traQgroups))
	for i, g := range traQgroups {
		groups[i] = formatV3Group(g)
	}
	return groups, err
}

func (repo *TraQRepository) GetUserBelongingGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	if repo.Version != TraQv1 {
		return nil, ErrForbidden
	}
	data, err := repo.getRequest(fmt.Sprintf("/users/%s/groups", userID))
	if err != nil {
		return nil, err
	}
	groupIDs := make([]uuid.UUID, 0)
	err = traQjson.Unmarshal(data, &groupIDs)
	return groupIDs, err
}

func (repo *TraPGroupRepository) CreateGroup(groupParams WriteGroupParams) (*Group, error) {
	return nil, ErrForbidden
}

func (repo *TraPGroupRepository) UpdateGroup(groupID uuid.UUID, groupParams WriteGroupParams) (*Group, error) {
	return nil, ErrForbidden
}

func (repo *TraPGroupRepository) AddUserToGroup(groupID uuid.UUID, userID uuid.UUID) error {
	return ErrForbidden
}

func (repo *TraPGroupRepository) DeleteGroup(groupID uuid.UUID) error {
	return ErrForbidden
}

func (repo *TraPGroupRepository) DeleteUserInGroup(groupID uuid.UUID, userID uuid.UUID) error {
	return ErrForbidden
}

func (repo *TraPGroupRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	if groupID != traPGroupID {
		return nil, ErrNotFound
	}
	return repo.getGroup()
}

func (repo *TraPGroupRepository) GetAllGroups() ([]*Group, error) {
	groups := make([]*Group, 0)
	group, err := repo.getGroup()
	groups = append(groups, group)
	return groups, err
}

func (repo *TraPGroupRepository) GetUserBelongingGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	group, err := repo.getGroup()
	if err != nil {
		return nil, err
	}
	if !group.IsMember(&GroupUsers{UserID: userID}) {
		return nil, ErrNotFound
	}
	return []uuid.UUID{traPGroupID}, nil
}

func formatV3Group(g *traQrouterV3.UserGroup) *Group {
	return &Group{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		JoinFreely:  false,
		Members:     formatV3GroupMemebers(g.Members),
		Admins:      formatV3GroupAdmins(g.Admins),
		CreatedBy:   g.Admins[0],
		Model: Model{
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		},
	}
}

func formatV3GroupMemebers(ms []traQrouterV3.UserGroupMember) []GroupUsers {
	users := make([]GroupUsers, len(ms))
	for i, m := range ms {
		users[i] = GroupUsers{
			UserID: m.ID,
		}
	}
	return users
}

func formatV3GroupAdmins(ms []uuid.UUID) []GroupAdmins {
	users := make([]GroupAdmins, len(ms))
	for i, m := range ms {
		users[i] = GroupAdmins{
			UserID: m,
		}
	}
	return users
}

func formatGroupMembers(groupID uuid.UUID, userIDs []uuid.UUID) []GroupUsers {
	users := make([]GroupUsers, len(userIDs))
	for i, v := range userIDs {
		users[i] = GroupUsers{
			GroupID: groupID,
			UserID:  v,
		}
	}
	return users
}

func formatGroupAdmins(groupID uuid.UUID, userIDs []uuid.UUID) []GroupAdmins {
	users := make([]GroupAdmins, len(userIDs))
	for i, v := range userIDs {
		users[i] = GroupAdmins{
			GroupID: groupID,
			UserID:  v,
		}
	}
	return users
}

func (g *Group) IsMember(user *GroupUsers) bool {
	for _, member := range g.Members {
		if member.UserID == user.UserID {
			return true
		}
	}
	return false
}

func (repo *TraPGroupRepository) getGroup() (*Group, error) {
	users, err := repo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	group := &Group{
		ID:          traPGroupID,
		Name:        "traP",
		Description: "traP全体グループ",
		JoinFreely:  false,
	}
	for _, user := range users {
		user := user
		group.Members = append(group.Members, GroupUsers{GroupID: group.ID, UserID: user.ID})
	}
	return group, nil
}

// BeforeCreate is gorm hook
func (g *Group) BeforeCreate() (err error) {
	g.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}
