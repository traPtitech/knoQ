package filter

import (
	"errors"
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

	// イベント開始時刻が指定された範囲内にあるか
	eventStartInRangeRight := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: LessEq,
		Value:    end,
	}
	eventStartInRangeLeft := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: GreterEq,
		Value:    start,
	}

	// イベント終了時刻が指定された範囲内にあるか
	eventEndInRangeRight := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: LessEq,
		Value:    end,
	}
	eventEndInRangeLeft := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: GreterEq,
		Value:    start,
	}

	// イベントの開催期間が指定された範囲を包含しているか
	eventStartBeforeRangeStart := &CmpExpr{
		Attr:     AttrTimeStart,
		Relation: LessEq,
		Value:    start,
	}
	eventEndAfterRangeEnd := &CmpExpr{
		Attr:     AttrTimeEnd,
		Relation: GreterEq,
		Value:    end,
	}

	return &LogicOpExpr{
		LogicOp: Or,
		Lhs: &LogicOpExpr{
			LogicOp: Or,
			Lhs: &LogicOpExpr{
				LogicOp: And,
				Lhs:     eventStartInRangeRight,
				Rhs:     eventStartInRangeLeft,
			},
			Rhs: &LogicOpExpr{
				LogicOp: And,
				Lhs:     eventEndInRangeRight,
				Rhs:     eventEndInRangeLeft,
			},
		},
		Rhs: &LogicOpExpr{
			LogicOp: And,
			Lhs:     eventStartBeforeRangeStart,
			Rhs:     eventEndAfterRangeEnd,
		},
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

// 以下のいずれかを満たす
// * 期間中に始まる (time_start >= since AND time_start < until)
// * 期間中に終わる (time_end >= since AND time_end < until)
// * 期間より前に始まり期間より後に終わる (time_start < since AND time_end >= until)
//
// ただしsinceがゼロのとき time_start <= until
// untilがゼロのとき since <= time_end
// 双方がゼロのき nil を返す
func FilterDuration(since, until time.Time) (Expr, error) {
	if since.IsZero() && until.IsZero() {
		return &CmpExpr{}, nil
	} else if since.IsZero() {
		return &CmpExpr{
			Attr:     AttrTimeStart,
			Relation: LessEq,
			Value:    until,
		}, nil
	} else if until.IsZero() {
		return &CmpExpr{
			Attr:     AttrTimeEnd,
			Relation: GreterEq,
			Value:    since,
		}, nil
	}

	if since.After(until) {
		return nil, errors.New("invalid time range")
	}

	// 期間中に始まる
	startIn := &LogicOpExpr{
		LogicOp: And,
		Lhs: &CmpExpr{
			Attr:     AttrTimeStart,
			Relation: GreterEq,
			Value:    since,
		},
		Rhs: &CmpExpr{
			Attr:     AttrTimeStart,
			Relation: Less,
			Value:    until,
		},
	}

	// 期間中に終わる
	endIn := &LogicOpExpr{
		LogicOp: And,
		Lhs: &CmpExpr{
			Attr:     AttrTimeEnd,
			Relation: GreterEq,
			Value:    since,
		},
		Rhs: &CmpExpr{
			Attr:     AttrTimeEnd,
			Relation: Less,
			Value:    until,
		},
	}

	// 期間より前に始まり期間より後に終わる
	throughout := &LogicOpExpr{
		LogicOp: And,
		Lhs: &CmpExpr{
			Attr:     AttrTimeStart,
			Relation: Less,
			Value:    since,
		},
		Rhs: &CmpExpr{
			Attr:     AttrTimeEnd,
			Relation: GreterEq,
			Value:    until,
		},
	}

	return &LogicOpExpr{
		LogicOp: Or,
		Lhs:     throughout,
		Rhs: &LogicOpExpr{
			LogicOp: Or,
			Lhs:     endIn,
			Rhs:     startIn,
		},
	}, nil
}
