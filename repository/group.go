package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	traQrouterV3 "github.com/traPtitech/traQ/router/v3"
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
	group.Members = formatGroupMembers(groupParams.Members)
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
	group.Members = formatGroupMembers(groupParams.Members)
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
		if !group.JoinFreely {
			return ErrForbidden
		}
		member := &User{ID: userID}
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
		if !group.JoinFreely {
			return ErrForbidden
		}
		member := &User{ID: userID}
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

func (repo *TraQRepository) getBaseURL() string {
	return traQEndPoints[repo.Version]
}

func (repo *TraQRepository) getRequest(path string) ([]byte, error) {
	if repo.Token == "" {
		return nil, ErrForbidden
	}
	req, err := http.NewRequest(http.MethodGet, repo.getBaseURL()+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+repo.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		// TODO consider 300
		switch res.StatusCode {
		case 401:
			return nil, ErrForbidden
		case 403:
			return nil, ErrForbidden
		case 404:
			return nil, ErrNotFound
		default:
			return nil, errors.New(http.StatusText(res.StatusCode))
		}
	}
	return ioutil.ReadAll(res.Body)

}

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
	if repo.Version != V3 {
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
	if repo.Version != V3 {
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
	if repo.Version != V1 {
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

func formatV3Group(g *traQrouterV3.UserGroup) *Group {
	return &Group{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		JoinFreely:  false,
		Members:     formatV3GroupMemebers(g.Members),
		CreatedBy:   g.Admins[0],
		Model: Model{
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		},
	}
}

func formatV3GroupMemebers(ms []traQrouterV3.UserGroupMember) []User {
	users := make([]User, len(ms))
	for i, m := range ms {
		users[i] = User{
			ID: m.ID,
		}
	}
	return users
}

func formatGroupMembers(userIDs []uuid.UUID) []User {
	users := make([]User, len(userIDs))
	for i, v := range userIDs {
		users[i] = User{
			ID: v,
		}
	}
	return users
}

func (g *Group) IsMember(user *User) bool {
	for _, member := range g.Members {
		if member.ID == user.ID {
			return true
		}
	}
	return false
}

// BeforeCreate is gorm hook
func (g *Group) BeforeCreate() (err error) {
	g.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}
