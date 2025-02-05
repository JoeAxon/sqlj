package sqlj

import "strings"

type QueryDB struct {
	DB           DB
	From         string
	WhereClauses []WhereClause
}

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
	return strings.Join([]string{"(", buildWhereClause(e.exprs), ")"}, "")
}

func (q QueryDB) Where(expr string) QueryDB {
	return q.WhereExpr(SimpleExpr{expr})
}

func (q QueryDB) WhereExpr(expr Expr) QueryDB {
	q.WhereClauses = append(q.WhereClauses, WhereClause{
		Type: "AND",
		Expr: expr,
	})

	return q
}

func (q QueryDB) OrWhere(expr string) QueryDB {
	return q.OrWhereExpr(SimpleExpr{expr})
}

func (q QueryDB) OrWhereExpr(expr Expr) QueryDB {
	q.WhereClauses = append(q.WhereClauses, WhereClause{
		Type: "OR",
		Expr: expr,
	})

	return q
}

func (q QueryDB) Get(id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := strings.Join([]string{"SELECT ", strings.Join(columns, ", "), " FROM ", q.From, " WHERE ", q.DB.GetIDName(), " = $1"}, "")

	return q.DB.GetRow(sql, v, id)
}

func (q QueryDB) One(v any, values ...any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := buildSelectQuery(columns, q.From, q.WhereClauses)

	return q.DB.GetRow(sql, v, values...)
}

func buildSelectQuery(columns []string, from string, whereClauses []WhereClause) string {
	sql := strings.Join([]string{"SELECT ", strings.Join(columns, ", "), " FROM ", from}, "")

	if len(whereClauses) > 0 {
		sql = strings.Join([]string{sql, " WHERE ", buildWhereClause(whereClauses)}, "")
	}

	return sql
}

func buildWhereClause(clauses []WhereClause) string {
	if len(clauses) == 0 {
		return ""
	}

	sql := make([]string, len(clauses)*2)

	for idx, clause := range clauses {
		if idx == 0 {
			sql[idx*2] = ""
		} else {
			sql[idx*2] = clause.Type
		}

		sql[idx*2+1] = clause.Expr.String()
	}

	return strings.Join(sql[1:], " ")
}

func replacePlaceholder(expr string, idx uint) (string, uint) {
	return expr, 0
}

func indexMatches(expr string) []uint {
	matches := []uint{}
	inQuote := false
	escaping := false

	for i := 0; i < len(expr); i++ {
		switch expr[i] {
		case '?':
			if !inQuote {
				matches = append(matches, uint(i))
			}
		case '\'':
			if !escaping {
				inQuote = !inQuote
			}
		}

		escaping = expr[i] == '\\'
	}

	return matches
}
