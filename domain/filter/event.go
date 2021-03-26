package filter

import (
	"time"

	"github.com/gofrs/uuid"
)

type Relation int

const (
	Eq Relation = iota
	Neq
	Greter
	Less
	GreterEq
	LessEq
)

type Attr int

const (
	Event Attr = iota
	Group
	Room
	User
	Tag
	Name
	TimeStart
	TimeEnd
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
	Attr     Attr
	Relation Relation
	Value    interface{}
}

func (LogicOpExpr) isExpr() {}
func (CmpExpr) isExpr()     {}

func FilterRoomIDs(roomIDs []uuid.UUID) Expr {
	if len(roomIDs) == 0 {
		return nil
	}

	var expr Expr
	expr = CmpExpr{
		Attr:     Room,
		Relation: Eq,
		Value:    roomIDs[0],
	}
	for _, id := range roomIDs[1:] {
		lhs := expr
		rhs := CmpExpr{
			Attr:     Room,
			Relation: Eq,
			Value:    id,
		}
		expr = LogicOpExpr{Or, lhs, rhs}
	}
	return expr
}

func FilterGroupIDs(groupIDs []uuid.UUID) Expr {
	if len(groupIDs) == 0 {
		return nil
	}

	var expr Expr
	expr = CmpExpr{
		Attr:     Group,
		Relation: Eq,
		Value:    groupIDs[0],
	}
	for _, id := range groupIDs[1:] {
		lhs := expr
		rhs := CmpExpr{
			Attr:     Group,
			Relation: Eq,
			Value:    id,
		}
		expr = LogicOpExpr{Or, lhs, rhs}
	}
	return expr
}

func FilterTime(start, end time.Time) Expr {
	if start.IsZero() && end.IsZero() {
		return nil
	}
	timeStart := CmpExpr{
		Attr:     TimeStart,
		Relation: GreterEq,
		Value:    start,
	}
	timeEnd := CmpExpr{
		Attr:     TimeEnd,
		Relation: LessEq,
		Value:    end,
	}

	if start.IsZero() {
		return timeEnd
	}
	if end.IsZero() {
		return timeStart
	}
	return LogicOpExpr{
		LogicOp: And,
		Lhs:     timeStart,
		Rhs:     timeEnd,
	}
}
