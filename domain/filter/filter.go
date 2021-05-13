package filter

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

func (*LogicOpExpr) isExpr() {}
func (*CmpExpr) isExpr()     {}
