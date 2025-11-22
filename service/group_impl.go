package service

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"gorm.io/gorm"
)

var traPGroupID = uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))

func (repo *service) CreateGroup(ctx context.Context, params domain.WriteGroupParams) (*domain.Group, error) {
	reqID, _ := domain.GetUserID(ctx)

	p := db.WriteGroupParams{
		WriteGroupParams: params,
		CreatedBy:        reqID,
	}
	g, err := repo.GormRepo.CreateGroup(p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	return g, nil
}

func (repo *service) UpdateGroup(ctx context.Context, groupID uuid.UUID, params domain.WriteGroupParams) (*domain.Group, error) {
	reqID, _ := domain.GetUserID(ctx)

	if !repo.IsGroupAdmins(ctx, groupID) {
		return nil, domain.ErrForbidden
	}
	p := db.WriteGroupParams{
		WriteGroupParams: params,
		CreatedBy:        reqID,
	}
	g, err := repo.GormRepo.UpdateGroup(groupID, p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	return g, nil
}

// AddMeToGroup add me to that group if that group is open.
func (repo *service) AddMeToGroup(ctx context.Context, groupID uuid.UUID) error {
	reqID, _ := domain.GetUserID(ctx)

	if !repo.IsGroupJoinFreely(ctx, groupID) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.AddMemberToGroup(groupID, reqID)
}

func (repo *service) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	if !repo.IsGroupAdmins(ctx, groupID) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.DeleteGroup(groupID)
}

// DeleteMeGroup delete me in that group if that group is open.
func (repo *service) DeleteMeGroup(ctx context.Context, groupID uuid.UUID) error {
	reqID, _ := domain.GetUserID(ctx)

	if !repo.IsGroupJoinFreely(ctx, groupID) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.DeleteMemberOfGroup(groupID, reqID)
}

func (repo *service) GetGroup(ctx context.Context, groupID uuid.UUID) (*domain.Group, error) {
	domainGroup, err := repo.GormRepo.GetGroup(groupID)
	if err == nil {
		return domainGroup, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, defaultErrorHandling(err)
	}

	// 以下 Not Found 用処理

	// traP全員用グループ
	if groupID == traPGroupID {
		return repo.getTraPGroup(ctx), nil
	}

	// traQ Group or error
	traQGroup, err := repo.TraQRepo.GetGroup(groupID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	group := ConvtraqUserGroupTodomainGroup(*traQGroup)
	group.IsTraQGroup = true

	return &group, nil
}

func (repo *service) GetAllGroups(ctx context.Context) ([]*domain.Group, error) {
	groups := make([]*domain.Group, 0)
	gg, err := repo.GormRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	groups = append(groups, gg...)
	tg, err := repo.TraQRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	dg := ConvSPtraqUserGroupToSPdomainGroup(tg)
	// add IsTraQGroup
	for i := range dg {
		dg[i].IsTraQGroup = true
	}
	// add trap
	groups = append(append(groups, repo.getTraPGroup(ctx)), dg...)

	return groups, nil
}

func (repo *service) GetUserBelongingGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	reqID, _ := domain.GetUserID(ctx)

	t, err := repo.GormRepo.GetToken(reqID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	ggIDs, err := repo.GormRepo.GetBelongGroupIDs(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	tgIDs, err := repo.TraQRepo.GetUserBelongingGroupIDs(t, userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	// add trap
	return append(append(ggIDs, traPGroupID), tgIDs...), nil
}

func (repo *service) GetUserAdminGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return repo.GormRepo.GetAdminGroupIDs(userID)
}

func (repo *service) IsGroupAdmins(ctx context.Context, groupID uuid.UUID) bool {
	reqID, _ := domain.GetUserID(ctx)

	group, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		return false
	}
	for _, admin := range group.Admins {
		if reqID == admin.ID {
			return true
		}
	}
	return false
}

func (repo *service) IsGroupJoinFreely(ctx context.Context, groupID uuid.UUID) bool {
	group, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		return false
	}
	return group.JoinFreely
}

func (repo *service) IsGroupMember(ctx context.Context, userID, groupID uuid.UUID) bool {
	group, err := repo.GetGroup(ctx, groupID)
	if err != nil {
		return false
	}
	for _, member := range group.Members {
		if userID == member.ID {
			return true
		}
	}

	return false
}

func (repo *service) getTraPGroup(ctx context.Context) *domain.Group {
	members, err := repo.GetAllUsers(ctx, false, false)
	if err != nil {
		return nil
	}

	return &domain.Group{
		ID:          traPGroupID,
		Name:        "traP",
		Description: "traP全体グループ",
		JoinFreely:  false,
		Members:     convSPdomainUserToSdomainUser(members),
		Admins:      []domain.User{},
		IsTraQGroup: true,
		CreatedBy:   domain.User{},
		Model:       domain.Model{},
	}
}

func (repo *service) GetGradeGroupNames(ctx context.Context) ([]string, error) {
	groups, err := repo.TraQRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// アクティブユーザーのみのmapを作成
	activeUsersID, err := repo.TraQRepo.GetUsers(false)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	activeUsersMap := make(map[string]struct{}, len(activeUsersID))
	for _, user := range activeUsersID {
		if user.Bot {
			continue
		}
		activeUsersMap[user.Id] = struct{}{}
	}

	names := make([]string, 0)
	for _, g := range groups {
		if g.Type == "grade" && len(g.Members) != 0 && g.Name != "00B" {
			for _, member := range g.Members {
				// グループメンバーが全員凍結されている場合は除外
				_, ok := activeUsersMap[member.Id]
				// アクティブメンバーが一人でもいる場合は追加
				if ok {
					names = append(names, g.Name)
					break
				}
			}
		}
	}
	return names, nil
}

func convSPdomainUserToSdomainUser(src []*domain.User) (dst []domain.User) {
	dst = make([]domain.User, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = (*src[i])
		}
	}
	return
}
