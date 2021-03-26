package production

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) CreateEvent(params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	// groupの確認
	_, err := repo.GetGroup(params.GroupID, info)
	if err != nil {
		return nil, err
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.gormRepo.CreateEvent(p)
	if err != nil {
		return nil, err
	}
	e := db.ConvertEventTodomainEvent(*event)
	return &e, nil
}

func (repo *Repository) UpdateEvent(eventID uuid.UUID, eventParams domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	panic("not implemented") // TODO: Implement
}

func (repo *Repository) AddTagToEvent(eventID uuid.UUID, tagID uuid.UUID, locked bool, info *domain.ConInfo) error {
	panic("not implemented") // TODO: Implement
}

func (repo *Repository) DeleteEvent(eventID uuid.UUID, info *domain.ConInfo) error {
	panic("not implemented") // TODO: Implement
}

// DeleteTagInEvent delete a tag in that Event
func (repo *Repository) DeleteTagInEvent(eventID uuid.UUID, tagID uuid.UUID, info *domain.ConInfo) error {
	panic("not implemented") // TODO: Implement
}

func (repo *Repository) GetEvent(eventID uuid.UUID) (*domain.Event, error) {
	panic("not implemented") // TODO: Implement
}

//go:generate gotypeconverter -s v3.UserGroup -d domain.Group -o converter.go .

func (repo *Repository) GetEvents(expr filter.Expr, info *domain.ConInfo) ([]*domain.Event, error) {
	expr = addTraQGroupIDs(repo, info.ReqUserID, expr)

	_, err := repo.gormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func addTraQGroupIDs(repo *Repository, userID uuid.UUID, expr filter.Expr) filter.Expr {
	t, err := repo.gormRepo.GetToken(userID)
	if err != nil {
		return expr
	}

	var fixExpr func(filter.Expr) filter.Expr

	fixExpr = func(expr filter.Expr) filter.Expr {
		switch e := expr.(type) {
		case *filter.CmpExpr:
			if e.Attr == filter.User {
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return e
				}
				groupIDs, err := repo.traQRepo.GetUserBelongingGroupIDs(t, id)
				if err != nil {
					return e
				}
				return &filter.LogicOpExpr{
					LogicOp: filter.Or,
					Lhs:     e,
					Rhs:     filter.FilterGroupIDs(groupIDs),
				}
			}
		case *filter.LogicOpExpr:
			return &filter.LogicOpExpr{
				LogicOp: e.LogicOp,
				Lhs:     fixExpr(e.Lhs),
				Rhs:     fixExpr(e.Rhs),
			}
		}
		return nil
	}
	return fixExpr(expr)
}
