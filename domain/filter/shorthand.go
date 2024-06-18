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

func addOr(lhs, rhs Expr) Expr {
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
		LogicOp: Or,
		Lhs:     lhs,
		Rhs:     rhs,
	}
}

// 期間中に始まる
// time_start >= min AND time_start < max
//
// 期間中に終わる
// time_end >= min AND time_end < max
//
// 期間より前に始まり期間より後に終わる
// time_start < min AND time_end >= max
//
// min < max であるべき
func FilterDuration(min, max time.Time) Expr {
	startAfterMin := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: GreterEq,
		Value:    min,
	}
	startBeforeMax := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: Less,
		Value:    max,
	}
	// 期間中に始まる
	startIn := &LogicOpExpr{
		LogicOp: And,
		Lhs:     startAfterMin,
		Rhs:     startBeforeMax,
	}

	endAfterMin := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: GreterEq,
		Value:    min,
	}
	endBeforeMax := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: Less,
		Value:    max,
	}
	// 期間中に終わる
	endIn := &LogicOpExpr{
		LogicOp: And,
		Lhs:     endAfterMin,
		Rhs:     endBeforeMax,
	}

	startBeforeMin := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: Less,
		Value:    min,
	}
	endAfterMax := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: GreterEq,
		Value:    max,
	}
	// 期間より前に始まり期間より後に終わる
	throughout := &LogicOpExpr{
		LogicOp: And,
		Lhs:     startBeforeMin,
		Rhs:     endAfterMax,
	}

	return addOr(addOr(startIn, endIn), throughout)
}
