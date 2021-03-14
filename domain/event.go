package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Event struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Room          Room
	Group         Group
	TimeStart     time.Time
	TimeEnd       time.Time
	CreatedBy     User
	Admins        []User
	Tags          []EventTag
	AllowTogether bool
	Model
}

type EventTag struct {
	Tag
	Locked bool
}

// for repository

// WriteEventParams is used create and update
type WriteEventParams struct {
	Name          string
	Description   string
	GroupID       uuid.UUID
	RoomID        uuid.UUID
	TimeStart     time.Time
	TimeEnd       time.Time
	Admins        []uuid.UUID
	AllowTogether bool
	Tags          []EventTagParams
}

type EventTagParams struct {
	Name   string
	Locked bool
}

// WriteTagRelationParams is used create and update
type WriteTagRelationParams struct {
	ID     uuid.UUID
	Locked bool
}

// EventRepository is implemented by ...
type EventRepository interface {
	CreateEvent(eventParams WriteEventParams, info *ConInfo) (*Event, error)

	UpdateEvent(eventID uuid.UUID, eventParams WriteEventParams, info *ConInfo) (*Event, error)
	AddTagToEvent(eventID uuid.UUID, tagID uuid.UUID, locked bool, info *ConInfo) error

	DeleteEvent(eventID uuid.UUID, info *ConInfo) error
	// DeleteTagInEvent delete a tag in that Event
	DeleteTagInEvent(eventID uuid.UUID, tagID uuid.UUID, info *ConInfo) error

	GetEvent(eventID uuid.UUID) (*Event, error)
	GetEvents(Expr) ([]*Event, error)

	GetEventActivities(day int) ([]*Event, error)
}

// Expression

func FilterRoomIDs(roomIDs []uuid.UUID) Expr {
	if len(roomIDs) == 0 {
		return nil
	}

	var expr Expr
	expr = RoomExpr{
		Relation: Eq,
		Value:    roomIDs[0],
	}
	for _, id := range roomIDs[1:] {
		lhs := expr
		rhs := RoomExpr{
			Relation: Eq,
			Value:    id,
		}
		expr = LogicOpExpr{Or, lhs, rhs}
	}

	return expr
}

type Relation int

const (
	Eq Relation = iota
	Neq
	Greter
	Less
	GreterEq
	LessEq
)

type LogicOp int

const (
	And LogicOp = iota
	Or
)

type Expr interface {
	isExpr()
}

type LogicOpExpr struct {
	LogicOp LogicOp
	Lhs     Expr
	Rhs     Expr
}

type CmpExpr struct {
	Attr     string
	Relation Relation
	Value    interface{}
}

type UserExpr struct {
	Relation Relation
	Value    uuid.UUID
}

type GroupExpr struct {
	Relation Relation
	Value    uuid.UUID
}

type RoomExpr struct {
	Relation Relation
	Value    uuid.UUID
}

type TagExpr struct {
	Relation Relation
	Value    uuid.UUID
}

type EventIDExpr struct {
	Relation Relation
	Value    uuid.UUID
}

type EventNameExpr struct {
	Relation Relation
	Value    string
}

type TimeStartExpr struct {
	Relation Relation
	Value    time.Time
}

type TimeEndExpr struct {
	Relation Relation
	Value    time.Time
}

type CmpInterface interface {
	Underlying() CmpExpr
}

func (cmp *CmpExpr) Underlying() CmpExpr {
	return *cmp
}

func (e *UserExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "user",
		Relation: e.Relation,
		Value:    e.Value,
	}
}

func (e *GroupExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "group",
		Relation: e.Relation,
		Value:    e.Value,
	}
}
func (e *RoomExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "room",
		Relation: e.Relation,
		Value:    e.Value,
	}
}
func (e *TagExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "tag",
		Relation: e.Relation,
		Value:    e.Value,
	}
}
func (e *EventIDExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "event",
		Relation: e.Relation,
		Value:    e.Value,
	}
}
func (e *EventNameExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "eventName",
		Relation: e.Relation,
		Value:    e.Value,
	}
}
func (e *TimeStartExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "start",
		Relation: e.Relation,
		Value:    e.Value,
	}
}
func (e *TimeEndExpr) Underlying() CmpExpr {
	return CmpExpr{
		Attr:     "end",
		Relation: e.Relation,
		Value:    e.Value,
	}
}

func (LogicOpExpr) isExpr() {}
func (CmpExpr) isExpr()     {}
func (UserExpr) isExpr()    {}
func (GroupExpr) isExpr()   {}
func (RoomExpr) isExpr()    {}
