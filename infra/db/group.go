package db

import (
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"github.com/traPtitech/knoQ/infra"
	"gorm.io/gorm"
)

func groupFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Members").Preload("Admins").Preload("CreatedBy")
}

type WriteGroupParams struct {
	domain.WriteGroupParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateGroup(params WriteGroupParams) (*domain.Group, error) {
	g, err := createGroup(repo.db, params)
	dg := ConvGroupTodomainGroup(*g)
	return &dg, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateGroup(groupID uuid.UUID, params WriteGroupParams) (*domain.Group, error) {
	g, err := updateGroup(repo.db, groupID, params)
	dg := ConvGroupTodomainGroup(*g)
	return &dg, defaultErrorHandling(err)
}

func (repo *GormRepository) AddMemberToGroup(groupID, userID uuid.UUID) error {
	err := repo.db.Model(&Group{ID: groupID}).Association("Members").Append(&User{ID: userID})
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteGroup(groupID uuid.UUID) error {
	err := deleteGroup(repo.db, groupID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteMemberOfGroup(groupID, userID uuid.UUID) error {
	err := repo.db.Model(&Group{ID: groupID}).Association("Members").Delete(&User{ID: userID})
	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetGroup(groupID uuid.UUID) (*domain.Group, error) {
	g, err := getGroup(groupFullPreload(repo.db), groupID)
	dg := ConvGroupTodomainGroup(*g)
	return &dg, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllGroups() ([]*domain.Group, error) {
	cmd := groupFullPreload(repo.db)
	gs, err := getGroups(cmd, "", nil)

	dgs := ConvSPGroupToSPdomainGroup(gs)
	return dgs, defaultErrorHandling(err)
}

func (repo *GormRepository) GetBelongGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	cmd := groupFullPreload(repo.db)
	filterFormat, filterArgs, err := createGroupFilter(filters.FilterBelongs(userID))
	if err != nil {
		return nil, err
	}
	gs, err := getGroups(cmd, filterFormat, filterArgs)

	return convSPGroupToSuuidUUID(gs), defaultErrorHandling(err)
}

func (repo *GormRepository) GetAdminGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	cmd := groupFullPreload(repo.db)
	filterFormat, filterArgs, err := createGroupFilter(filters.FilterAdmins(userID))
	if err != nil {
		return nil, err
	}
	gs, err := getGroups(cmd, filterFormat, filterArgs)

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

func deleteGroup(db *gorm.DB, groupID uuid.UUID) error {
	group := Group{
		ID: groupID,
	}
	return db.Delete(&group).Error
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
		filters.AttrUser:   "group_member.user_id",
		filters.AttrBelong: "group_member.user_id",
		filters.AttrAdmin:  "group_admin.user_id",

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

func (repo *GormRepository) SyncExternalGroups() error {
	externalGroups, err := repo.traqRepo.GetAllGroups() // 外部APIから取得
	if err != nil {
		return err
	}

	existingTraqGroups := make([]*Group, 0)
	err = repo.db.Model(&Group{IsTraqGroup: true}).Find(&existingTraqGroups).Error
	if err != nil {
		return err
	}

	existingTraqGroupsMap := make(map[uuid.UUID]*Group, 0)
	for _, g := range existingTraqGroups {
		existingTraqGroupsMap[g.TraqID.UUID] = g
	}

	newGroups := make([]*Group, 0, len(externalGroups))
	for _, g := range externalGroups {
		if g.ID.IsNil() {
			continue
		}
		var gid uuid.UUID
		if _, ok := existingTraqGroupsMap[g.ID]; ok {
			gid = g.ID
		} else {
			gid, err = uuid.NewV4()
			if err != nil {
				return err
			}
		}

		newGroups = append(newGroups, &Group{
			ID:          gid,
			TraqID:      uuid.NullUUID{UUID: g.ID, Valid: true},
			Name:        g.Name,
			Description: g.Description,
			IsTraqGroup: true,
			JoinFreely:  sql.NullBool{},
			Members: lo.Map(g.Members, func(tm infra.TraqUserGroupMember, _ int) *User {
				return &User{
					ID: tm.ID,
				}
			}),
			Admins: lo.Map(g.Admins, func(id uuid.UUID, _ int) *User {
				return &User{
					ID: id,
				}
			}),
			CreatedByRefer: uuid.NullUUID{},
		})
	}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		existingGroups := []Group{}
		if err := tx.Where("is_traq_group = ?", true).Find(&existingGroups).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(&newGroups).Error; err != nil {
			return err
		}

		return nil
	})
}
