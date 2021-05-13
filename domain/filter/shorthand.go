package filter

import (
	"time"

	"github.com/gofrs/uuid"
)

func FilterRoomIDs(roomIDs ...uuid.UUID) Expr {
	if len(roomIDs) == 0 {
		return nil
	}

	var expr Expr
	expr = &CmpExpr{
		Attr:     AttrRoom,
		Relation: Eq,
		Value:    roomIDs[0],
	}
	for _, id := range roomIDs[1:] {
		lhs := expr
		rhs := &CmpExpr{
			Attr:     AttrRoom,
			Relation: Eq,
			Value:    id,
		}
		expr = &LogicOpExpr{Or, lhs, rhs}
	}
	return expr
}

func FilterGroupIDs(groupIDs ...uuid.UUID) Expr {
	if len(groupIDs) == 0 {
		return nil
	}

	var expr Expr
	expr = &CmpExpr{
		Attr:     AttrGroup,
		Relation: Eq,
		Value:    groupIDs[0],
	}
	for _, id := range groupIDs[1:] {
		lhs := expr
		rhs := &CmpExpr{
			Attr:     AttrGroup,
			Relation: Eq,
			Value:    id,
		}
		expr = &LogicOpExpr{Or, lhs, rhs}
	}
	return expr
}

func FilterUserIDs(userIDs ...uuid.UUID) Expr {
	if len(userIDs) == 0 {
		return nil
	}

	var expr Expr
	expr = &CmpExpr{
		Attr:     AttrUser,
		Relation: Eq,
		Value:    userIDs[0],
	}
	for _, id := range userIDs[1:] {
		lhs := expr
		rhs := &CmpExpr{
			Attr:     AttrUser,
			Relation: Eq,
			Value:    id,
		}
		expr = &LogicOpExpr{Or, lhs, rhs}
	}
	return expr
}

func FilterTime(start, end time.Time) Expr {
	if start.IsZero() && end.IsZero() {
		return nil
	}
	timeStart := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: GreterEq,
		Value:    start,
	}
	timeEnd := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: LessEq,
		Value:    end,
	}

	if start.IsZero() {
		return timeEnd
	}
	if end.IsZero() {
		return timeStart
	}
	return &LogicOpExpr{
		LogicOp: And,
		Lhs:     timeStart,
		Rhs:     timeEnd,
	}
}

func AddAnd(lhs, rhs Expr) Expr {
	if lhs == nil && rhs == nil {
		return nil
	}
	if lhs == nil {
		return rhs
	}
	if rhs == nil {
		return lhs
	}
	return &LogicOpExpr{
		LogicOp: And,
		Lhs:     lhs,
		Rhs:     rhs,
	}
}
