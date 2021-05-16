package db

import (
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"gorm.io/gorm"
)

func groupFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Members").Preload("Admins").Preload("CreatedBy")
}

type WriteGroupParams struct {
	domain.WriteGroupParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateGroup(params WriteGroupParams) (*Group, error) {
	g, err := createGroup(repo.db, params)
	return g, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateGroup(groupID uuid.UUID, params WriteGroupParams) (*Group, error) {
	g, err := updateGroup(repo.db, groupID, params)
	return g, defaultErrorHandling(err)
}

func (repo *GormRepository) AddMemberToGroup(groupID, userID uuid.UUID) error {
	err := addMemberToGroup(repo.db, groupID, userID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteGroup(groupID uuid.UUID) error {
	err := deleteGroup(repo.db, groupID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteMemberOfGroup(groupID, userID uuid.UUID) error {
	err := deleteMemberOfGroup(repo.db, groupID, userID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	gs, err := getGroup(groupFullPreload(repo.db), groupID)
	return gs, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllGroups() ([]*Group, error) {
	cmd := groupFullPreload(repo.db)
	gs, err := getGroups(cmd.Joins(
		"LEFT JOIN events ON groups.id = events.group_id "+
			"LEFT JOIN group_members ON groups.id = group_members.group_id "+
			"LEFT JOIN group_admins On groups.id = group_admins.group_id "), "", nil)
	return gs, defaultErrorHandling(err)
}

func (repo *GormRepository) GetBelongGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	cmd := groupFullPreload(repo.db)
	filterFormat, filterArgs, err := createGroupFilter(filter.FilterBelongs(userID))
	if err != nil {
		return nil, err
	}
	gs, err := getGroups(cmd.Joins(
		"LEFT JOIN events ON groups.id = events.group_id "+
			"LEFT JOIN group_members ON groups.id = group_members.group_id "+
			"LEFT JOIN group_admins On groups.id = group_admins.group_id "), filterFormat, filterArgs)

	return convSPGroupToSuuidUUID(gs), defaultErrorHandling(err)
}

func (repo *GormRepository) GetAdminGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	cmd := groupFullPreload(repo.db)
	filterFormat, filterArgs, err := createGroupFilter(filter.FilterAdmins(userID))
	if err != nil {
		return nil, err
	}
	gs, err := getGroups(cmd.Joins(
		"LEFT JOIN events ON groups.id = events.group_id "+
			"LEFT JOIN group_members ON groups.id = group_members.group_id "+
			"LEFT JOIN group_admins On groups.id = group_admins.group_id "), filterFormat, filterArgs)

	return convSPGroupToSuuidUUID(gs), defaultErrorHandling(err)
}

func convSPGroupToSuuidUUID(src []*Group) (dst []uuid.UUID) {
	dst = make([]uuid.UUID, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = (*src[i]).ID
		}
	}
	return
}

func createGroup(db *gorm.DB, groupParams WriteGroupParams) (*Group, error) {
	group := ConvWriteGroupParamsToGroup(groupParams)
	err := db.Create(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func updateGroup(db *gorm.DB, groupID uuid.UUID, params WriteGroupParams) (*Group, error) {
	group := ConvWriteGroupParamsToGroup(params)
	group.ID = groupID
	err := db.Session(&gorm.Session{FullSaveAssociations: true}).
		Omit("CreatedAt").Save(&group).Error
	return &group, err
}

func addMemberToGroup(db *gorm.DB, groupID, userID uuid.UUID) error {
	groupMember := GroupMember{
		GroupID: groupID,
		UserID:  userID,
	}
	err := db.Create(&groupMember).Error
	if errors.Is(defaultErrorHandling(err), ErrDuplicateEntry) {
		return db.Omit("CreatedAt").Save(&groupMember).Error
	}
	return err
}

func deleteGroup(db *gorm.DB, groupID uuid.UUID) error {
	group := Group{
		ID: groupID,
	}
	return db.Delete(&group).Error
}

func deleteMemberOfGroup(db *gorm.DB, groupID, userID uuid.UUID) error {
	groupMember := GroupMember{
		GroupID: groupID,
		UserID:  userID,
	}
	return db.Delete(&groupMember).Error
}

func getGroup(db *gorm.DB, groupID uuid.UUID) (*Group, error) {
	group := Group{}
	err := db.Take(&group, groupID).Error
	return &group, err
}

func getGroups(db *gorm.DB, query string, args []interface{}) ([]*Group, error) {
	groups := make([]*Group, 0)
	cmd := db
	if query != "" && args != nil {
		cmd = cmd.Where(query, args)
	}
	err := cmd.Group("id").Find(&groups).Error
	return groups, err
}

func createGroupFilter(expr filter.Expr) (string, []interface{}, error) {
	if expr == nil {
		return "", []interface{}{}, nil
	}

	attrMap := map[filter.Attr]string{
		filter.AttrUser:   "group_members.user_id",
		filter.AttrBelong: "group_members.user_id",
		filter.AttrAdmin:  "group_admins.user_id",

		filter.AttrName:  "groups.name",
		filter.AttrGroup: "groups.id",
		filter.AttrEvent: "events.id",
	}
	defaultRelationMap := map[filter.Relation]string{
		filter.Eq:       "=",
		filter.Neq:      "!=",
		filter.Greter:   ">",
		filter.GreterEq: ">=",
		filter.Less:     "<",
		filter.LessEq:   "<=",
	}

	var cf func(e filter.Expr) (string, []interface{}, error)
	cf = func(e filter.Expr) (string, []interface{}, error) {
		var filterFormat string
		var filterArgs []interface{}

		switch e := e.(type) {
		case *filter.CmpExpr:
			switch e.Attr {
			case filter.AttrName:
				name, ok := e.Value.(string)
				if !ok {
					return "", nil, ErrExpression
				}
				rel := map[filter.Relation]string{
					filter.Eq:  "=",
					filter.Neq: "!=",
				}[e.Relation]
				filterFormat = fmt.Sprintf("groups.name %v ?", rel)
				filterArgs = []interface{}{name}
			default:
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return "", nil, ErrExpression
				}
				filterFormat = fmt.Sprintf("%v %v ?", attrMap[e.Attr], defaultRelationMap[e.Relation])
				filterArgs = []interface{}{id}
			}

		case *filter.LogicOpExpr:
			op := map[filter.LogicOp]string{
				filter.And: "AND",
				filter.Or:  "OR",
			}[e.LogicOp]
			lFilter, lFilterArgs, lerr := cf(e.Lhs)
			rFilter, rFilterArgs, rerr := cf(e.Rhs)

			if lerr != nil && rerr != nil {
				return "", nil, ErrExpression
			}
			if lerr != nil {
				return rFilter, rFilterArgs, nil
			}
			if rerr != nil {
				return lFilter, lFilterArgs, nil
			}

			filterFormat = fmt.Sprintf("( %v ) %v ( %v )", lFilter, op, rFilter)
			filterArgs = append(lFilterArgs, rFilterArgs...)

		default:
			return "", nil, ErrExpression
		}

		return filterFormat, filterArgs, nil
	}
	return cf(expr)
}
