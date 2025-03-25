package filters

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
	AttrEvent Attr = iota
	AttrGroup
	AttrRoom
	AttrUser
	AttrTag
	AttrName
	AttrTimeStart
	AttrTimeEnd
	AttrAdmin
	AttrBelong
	AttrAttendee
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

func (*LogicOpExpr) isExpr() {}
func (*CmpExpr) isExpr()     {}
