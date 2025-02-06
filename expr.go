package sqlj

import "errors"

type QueryDB struct {
	DB           *DB
	From         string
	OrderClauses []OrderBy
	WhereClauses []WhereClause
}

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

func (q QueryDB) Where(expr string) QueryDB {
	return q.WhereExpr(SimpleExpr{expr})
}

func (q QueryDB) WhereExpr(expr Expr) QueryDB {
	q.WhereClauses = append(q.WhereClauses, WhereClause{
		Type: AND_TYPE,
		Expr: expr,
	})

	return q
}

func (q QueryDB) OrWhere(expr string) QueryDB {
	return q.OrWhereExpr(SimpleExpr{expr})
}

func (q QueryDB) OrWhereExpr(expr Expr) QueryDB {
	q.WhereClauses = append(q.WhereClauses, WhereClause{
		Type: OR_TYPE,
		Expr: expr,
	})

	return q
}

func (q QueryDB) Order(expression string, direction string) QueryDB {
	q.OrderClauses = append(q.OrderClauses, OrderBy{
		expression: expression,
		direction:  direction,
	})

	return q
}

// Get a record by ID.
// This will ignore any previous calls to .Where and .OrWhere
func (q QueryDB) Get(id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := buildSelectQuery(Select{
		Columns: columns,
		From:    q.From,
		Where: []WhereClause{
			{AND_TYPE, SimpleExpr{columnEq(q.DB.getIDName())}},
		},
	})

	return q.DB.GetRow(sql, v, id)
}

// Get a single record from the given table.
func (q QueryDB) One(v any, values ...any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := buildSelectQuery(Select{
		Columns: columns,
		From:    q.From,
		Where:   q.WhereClauses,
	})

	return q.DB.GetRow(sql, v, values...)
}

func (q QueryDB) Select(v any, values ...any) error {
	structInstance, err := getSliceStructInstance(v)
	if err != nil {
		return err
	}

	fields := extractFields(structInstance)
	columns := pluckNames(fields)

	sql := buildSelectQuery(Select{
		From:    q.From,
		Where:   q.WhereClauses,
		OrderBy: q.OrderClauses,
		Columns: columns,
	})

	return q.DB.SelectAll(sql, v)
}

func (q QueryDB) Page(options PageOptions, v any, values ...any) error {
	if options.pageNumber < 1 {
		return errors.New("Page number must be greater than 0")
	}

	if options.pageSize < 1 {
		return errors.New("Page size must be greater than 0")
	}

	if len(options.order) == 0 {
		return errors.New("Must include atleast one order by")
	}

	structInstance, err := getSliceStructInstance(v)
	if err != nil {
		return err
	}

	fields := extractFields(structInstance)
	columns := pluckNames(fields)

	offset := (options.pageNumber - 1) * options.pageSize
	limit := options.pageSize

	sql := buildSelectQuery(Select{
		From:    q.From,
		Where:   q.WhereClauses,
		OrderBy: options.order,
		Columns: columns,
		Offset:  true,
		Limit:   true,
	})

	return q.DB.SelectAll(sql, v, offset, limit)
}
