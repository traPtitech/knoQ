package service

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

var traPGroupID = uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))

func (s *service) CreateGroup(ctx context.Context, reqID uuid.UUID, params domain.WriteGroupParams) (*domain.Group, error) {

	p := domain.UpsertGroupArgs{
		WriteGroupParams: params,
		CreatedBy:        reqID,
	}
	g, err := s.GormRepo.CreateGroup(p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	return g, nil
}

func (s *service) UpdateGroup(ctx context.Context, reqID uuid.UUID, groupID uuid.UUID, params domain.WriteGroupParams) (*domain.Group, error) {
	if !s.IsGroupAdmins(ctx, reqID, groupID) {
		return nil, domain.ErrForbidden
	}
	p := domain.UpsertGroupArgs{
		WriteGroupParams: params,
		CreatedBy:        reqID,
	}
	g, err := s.GormRepo.UpdateGroup(groupID, p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	return g, nil
}

// AddMeToGroup add me to that group if that group is open.
func (s *service) AddMeToGroup(ctx context.Context, reqID uuid.UUID, groupID uuid.UUID) error {

	if !s.IsGroupJoinFreely(ctx, groupID) {
		return domain.ErrForbidden
	}
	return s.GormRepo.AddMemberToGroup(groupID, reqID)
}

func (s *service) DeleteGroup(ctx context.Context, reqID uuid.UUID, groupID uuid.UUID) error {
	if !s.IsGroupAdmins(ctx, reqID, groupID) {
		return domain.ErrForbidden
	}
	return s.GormRepo.DeleteGroup(groupID)
}

// DeleteMeGroup delete me in that group if that group is open.
func (s *service) DeleteMeGroup(ctx context.Context, reqID uuid.UUID, groupID uuid.UUID) error {
	if !s.IsGroupJoinFreely(ctx, groupID) {
		return domain.ErrForbidden
	}
	return s.GormRepo.DeleteMemberOfGroup(groupID, reqID)
}

func (s *service) GetGroup(ctx context.Context, groupID uuid.UUID) (*domain.Group, error) {
	domainGroup, err := s.GormRepo.GetGroup(groupID)
	if err == nil {
		return domainGroup, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, defaultErrorHandling(err)
	}

	// 以下 Not Found 用処理

	// traP全員用グループ
	if groupID == traPGroupID {
		return s.getTraPGroup(ctx), nil
	}

	// traQ Group or error
	traQGroup, err := s.TraQRepo.GetGroup(groupID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	group := ConvtraqUserGroupTodomainGroup(*traQGroup)
	group.IsTraQGroup = true

	return &group, nil
}

func (s *service) GetAllGroups(ctx context.Context) ([]*domain.Group, error) {
	groups := make([]*domain.Group, 0)
	gg, err := s.GormRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	groups = append(groups, gg...)
	tg, err := s.TraQRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	dg := ConvSPtraqUserGroupToSPdomainGroup(tg)
	// add IsTraQGroup
	for i := range dg {
		dg[i].IsTraQGroup = true
	}
	// add trap
	groups = append(append(groups, s.getTraPGroup(ctx)), dg...)

	return groups, nil
}

func (s *service) GetUserBelongingGroupIDs(ctx context.Context, reqID uuid.UUID, userID uuid.UUID) ([]uuid.UUID, error) {

	t, err := s.GormRepo.GetToken(reqID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	ggIDs, err := s.GormRepo.GetBelongGroupIDs(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	tgIDs, err := s.TraQRepo.GetUserBelongingGroupIDs(t, userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	// add trap
	return append(append(ggIDs, traPGroupID), tgIDs...), nil
}

func (s *service) GetUserAdminGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.GormRepo.GetAdminGroupIDs(userID)
}

func (s *service) IsGroupAdmins(ctx context.Context, reqID uuid.UUID, groupID uuid.UUID) bool {
	group, err := s.GormRepo.GetGroup(groupID)
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

func (s *service) IsGroupJoinFreely(ctx context.Context, groupID uuid.UUID) bool {
	group, err := s.GormRepo.GetGroup(groupID)
	if err != nil {
		return false
	}
	return group.JoinFreely
}

func (s *service) IsGroupMember(ctx context.Context, userID, groupID uuid.UUID) bool {
	group, err := s.GetGroup(ctx, groupID)
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

func (s *service) getTraPGroup(ctx context.Context) *domain.Group {
	members, err := s.GetAllUsers(ctx, false, false)
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

func (s *service) GetGradeGroupNames(ctx context.Context) ([]string, error) {
	groups, err := s.TraQRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// アクティブユーザーのみのmapを作成
	activeUsersID, err := s.TraQRepo.GetUsers(false)
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
