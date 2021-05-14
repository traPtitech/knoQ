package filter

import (
	"time"

	"github.com/gofrs/uuid"
)

func filterIDs(attr Attr, ids []uuid.UUID) Expr {
	if len(ids) == 0 {
		return nil
	}

	var expr Expr
	expr = &CmpExpr{
		Attr:     attr,
		Relation: Eq,
		Value:    ids[0],
	}
	for _, id := range ids[1:] {
		lhs := expr
		rhs := &CmpExpr{
			Attr:     attr,
			Relation: Eq,
			Value:    id,
		}
		expr = &LogicOpExpr{Or, lhs, rhs}
	}
	return expr
}

func FilterRoomIDs(roomIDs ...uuid.UUID) Expr {
	return filterIDs(AttrRoom, roomIDs)
}

func FilterGroupIDs(groupIDs ...uuid.UUID) Expr {
	return filterIDs(AttrGroup, groupIDs)
}

func FilterUserIDs(userIDs ...uuid.UUID) Expr {
	return filterIDs(AttrUser, userIDs)
}

func FilterBelongs(userIDs ...uuid.UUID) Expr {
	return filterIDs(AttrBelong, userIDs)
}

func FilterAdmins(userIDs ...uuid.UUID) Expr {
	return filterIDs(AttrAdmin, userIDs)
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
