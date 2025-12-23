package db

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func groupFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Members").Preload("Admins").Preload("CreatedBy")
}

func (repo *gormRepository) CreateGroup(args domain.UpsertGroupArgs) (*domain.Group, error) {
	g, err := createGroup(repo.db, args)
	domainGroup := convGroupTodomainGroup(*g)
	return &domainGroup, defaultErrorHandling(err)
}

func (repo *gormRepository) UpdateGroup(groupID uuid.UUID, args domain.UpsertGroupArgs) (*domain.Group, error) {
	g, err := updateGroup(repo.db, groupID, args)
	domainGroup := convGroupTodomainGroup(*g)
	return &domainGroup, defaultErrorHandling(err)
}

func (repo *gormRepository) AddMemberToGroup(groupID, userID uuid.UUID) error {
	err := addMemberToGroup(repo.db, groupID, userID)
	return defaultErrorHandling(err)
}

func (repo *gormRepository) DeleteGroup(groupID uuid.UUID) error {
	err := deleteGroup(repo.db, groupID)
	return defaultErrorHandling(err)
}

func (repo *gormRepository) DeleteMemberOfGroup(groupID, userID uuid.UUID) error {
	err := deleteMemberOfGroup(repo.db, groupID, userID)
	return defaultErrorHandling(err)
}

func (repo *gormRepository) GetGroup(groupID uuid.UUID) (*domain.Group, error) {
	g, err := getGroup(groupFullPreload(repo.db), groupID)
	domainGroup := convGroupTodomainGroup(*g)
	return &domainGroup, defaultErrorHandling(err)
}

func (repo *gormRepository) GetAllGroups() ([]*domain.Group, error) {
	cmd := groupFullPreload(repo.db)
	gs, err := getGroups(cmd.Joins(
		"LEFT JOIN events ON groups.id = events.group_id "+
			"LEFT JOIN group_members ON groups.id = group_members.group_id "+
			"LEFT JOIN group_admins On groups.id = group_admins.group_id "), "", nil)

	return ConvSPGroupToSPdomainGroup(gs), defaultErrorHandling(err)
}

func (repo *gormRepository) GetBelongGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	cmd := groupFullPreload(repo.db)
	filterFormat, filterArgs, err := createGroupFilter(filters.FilterBelongs(userID))
	if err != nil {
		return nil, err
	}
	gs, err := getGroups(cmd.Joins(
		"LEFT JOIN events ON groups.id = events.group_id "+
			"LEFT JOIN group_members ON groups.id = group_members.group_id "+
			"LEFT JOIN group_admins On groups.id = group_admins.group_id "), filterFormat, filterArgs)

	return convSPGroupToSuuidUUID(gs), defaultErrorHandling(err)
}

func (repo *gormRepository) GetAdminGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	cmd := groupFullPreload(repo.db)
	filterFormat, filterArgs, err := createGroupFilter(filters.FilterAdmins(userID))
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
			dst[i] = src[i].ID
		}
	}
	return
}

func createGroup(db *gorm.DB, args domain.UpsertGroupArgs) (*Group, error) {
	group := ConvWriteGroupParamsToGroup(args)
	err := db.Create(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func updateGroup(db *gorm.DB, groupID uuid.UUID, args domain.UpsertGroupArgs) (*Group, error) {
	group := ConvWriteGroupParamsToGroup(args)
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

	onConflictClause := clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "deleted_at"}),
	}

	return db.Clauses(onConflictClause).Create(&groupMember).Error
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

func createGroupFilter(expr filters.Expr) (string, []interface{}, error) {
	if expr == nil {
		return "", []interface{}{}, nil
	}

	attrMap := map[filters.Attr]string{
		filters.AttrUser:   "group_members.user_id",
		filters.AttrBelong: "group_members.user_id",
		filters.AttrAdmin:  "group_admins.user_id",

		filters.AttrName:  "groups.name",
		filters.AttrGroup: "groups.id",
		filters.AttrEvent: "events.id",
	}
	defaultRelationMap := map[filters.Relation]string{
		filters.Eq:       "=",
		filters.Neq:      "!=",
		filters.Greter:   ">",
		filters.GreterEq: ">=",
		filters.Less:     "<",
		filters.LessEq:   "<=",
	}

	var cf func(e filters.Expr) (string, []interface{}, error)
	cf = func(e filters.Expr) (string, []interface{}, error) {
		var filterFormat string
		var filterArgs []interface{}

		switch e := e.(type) {
		case *filters.CmpExpr:
			switch e.Attr {
			case filters.AttrName:
				name, ok := e.Value.(string)
				if !ok {
					return "", nil, ErrExpression
				}
				rel := map[filters.Relation]string{
					filters.Eq:  "=",
					filters.Neq: "!=",
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

		case *filters.LogicOpExpr:
			op := map[filters.LogicOp]string{
				filters.And: "AND",
				filters.Or:  "OR",
			}[e.LogicOp]
			lFilter, lFilterArgs, lerr := cf(e.LHS)
			rFilter, rFilterArgs, rerr := cf(e.RHS)

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
			filterArgs = lFilterArgs
			filterArgs = append(filterArgs, rFilterArgs...)

		default:
			return "", nil, ErrExpression
		}

		return filterFormat, filterArgs, nil
	}
	return cf(expr)
}
