package sqlj

import "errors"

type QueryDB struct {
	DB           *DB
	From         string
	OrderClauses []orderBy
	WhereClauses []WhereClause
	WhereValues  []any
}

func (q QueryDB) Where(expr string, values ...any) QueryDB {
	return q.WhereExpr(SimpleExpr{expr}, values...)
}

func (q QueryDB) WhereExpr(expr Expr, values ...any) QueryDB {
	q.WhereClauses = append(q.WhereClauses, WhereClause{
		Type: AND_TYPE,
		Expr: expr,
	})

	q.WhereValues = append(q.WhereValues, values...)

	return q
}

func (q QueryDB) OrWhere(expr string, values ...any) QueryDB {
	return q.OrWhereExpr(SimpleExpr{expr})
}

func (q QueryDB) OrWhereExpr(expr Expr, values ...any) QueryDB {
	q.WhereClauses = append(q.WhereClauses, WhereClause{
		Type: OR_TYPE,
		Expr: expr,
	})

	q.WhereValues = append(q.WhereValues, values...)

	return q
}

func (q QueryDB) Order(expression string, direction string) QueryDB {
	q.OrderClauses = append(q.OrderClauses, orderBy{
		Expression: expression,
		Direction:  direction,
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

	sql := buildSelectQuery(selectParams{
		Columns: columns,
		From:    q.From,
		Where: []WhereClause{
			{AND_TYPE, SimpleExpr{columnEq(q.DB.getIDName())}},
		},
	})

	return q.DB.GetRow(sql, v, id)
}

// Get a single record from the given table.
func (q QueryDB) One(v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := buildSelectQuery(selectParams{
		Columns: columns,
		From:    q.From,
		Where:   q.WhereClauses,
	})

	return q.DB.GetRow(sql, v, q.WhereValues...)
}

// Select all data from the query object.
// The results will be marshalled into the v slice of structs.
// v must be a pointer to a slice of structs.
func (q QueryDB) All(v any) error {
	structInstance, err := getSliceStructInstance(v)
	if err != nil {
		return err
	}

	fields := extractFields(structInstance)
	columns := pluckNames(fields)

	sql := buildSelectQuery(selectParams{
		From:    q.From,
		Where:   q.WhereClauses,
		OrderBy: q.OrderClauses,
		Columns: columns,
	})

	return q.DB.SelectAll(sql, v, q.WhereValues...)
}

// Selects a page of data from the given table.
// The options parameter allows you to specify the page and page size.
// The results will be marshalled into the v slice of structs.
// v must be a pointer to a slice of structs.
func (q QueryDB) Page(page uint, pageSize uint, v any) error {
	if page < 1 {
		return errors.New("Page number must be greater than 0")
	}

	if pageSize < 1 {
		return errors.New("Page size must be greater than 0")
	}

	structInstance, err := getSliceStructInstance(v)
	if err != nil {
		return err
	}

	fields := extractFields(structInstance)
	columns := pluckNames(fields)

	offset := (page - 1) * pageSize
	limit := pageSize

	sql := buildSelectQuery(selectParams{
		From:    q.From,
		Where:   q.WhereClauses,
		OrderBy: q.OrderClauses,
		Columns: columns,
		Offset:  true,
		Limit:   true,
	})

	values := append(q.WhereValues, limit, offset)

	return q.DB.SelectAll(sql, v, values...)
}

// Counts the number of records in the table.
// This is intended to be used in conjunction with .Page.
func (q QueryDB) Count() (uint, error) {
	var count uint = 0

	sql := buildSelectQuery(selectParams{
		From:    q.From,
		Where:   q.WhereClauses,
		Columns: []string{"count(1)"},
	})

	result := q.DB.DB.QueryRow(sql, q.WhereValues...)

	if err := result.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
