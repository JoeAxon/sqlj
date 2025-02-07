package sqlj

const (
	AND_TYPE = "AND"
	OR_TYPE  = "OR"
)

type WhereClause struct {
	Type string
	Expr Expr
}

type Expr interface {
	String() string
}

type SimpleExpr struct {
	expr string
}

type NestedExpr struct {
	exprs []WhereClause
}

func (e SimpleExpr) String() string {
	return e.expr
}

func (e NestedExpr) String() string {
	return parens(joinWhereClauses(e.exprs))
}
