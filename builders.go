package sqlj

import (
	"fmt"
	"slices"
	"strings"
)

type deleteParams struct {
	From  string
	Where []WhereClause
}

func buildDeleteSQL(options deleteParams) string {
	sql := strings.Join([]string{"DELETE FROM ", options.From}, "")

	if len(options.Where) > 0 {
		whereSQL, _ := buildWhereClause(options.Where)

		sql = strings.Join([]string{sql, " WHERE ", whereSQL}, "")
	}

	return sql
}

type insertParams struct {
	From      string
	Fields    []Field
	Returning []string
}

func buildInsertSQL(options insertParams) string {
	names := make([]string, len(options.Fields))
	placeholders := make([]string, len(options.Fields))

	n := 0
	for idx, f := range options.Fields {
		names[idx] = f.GetName()
		placeholders[idx] = f.GetPlaceholder(n + 1)

		if !f.IsLiteral() {
			n++
		}
	}

	return strings.Join(
		[]string{
			"INSERT INTO ",
			options.From,
			" (",
			strings.Join(names, ", "),
			") VALUES (",
			strings.Join(placeholders, ", "),
			") RETURNING ",
			strings.Join(options.Returning, ", "),
		},
		"",
	)
}

type updateParams struct {
	From      string
	Fields    []Field
	Returning []string
}

func buildUpdateSQL(options updateParams) string {
	setExpressions := make([]string, len(options.Fields))
	for idx, f := range options.Fields {
		setExpressions[idx] = fmt.Sprintf("%s = %s", f.GetName(), f.GetPlaceholder(idx+1))
	}

	return strings.Join(
		[]string{
			"UPDATE ",
			options.From,
			" SET ",
			strings.Join(setExpressions, ", "),
			fmt.Sprintf(" WHERE id = $%d ", len(options.Fields)+1),
			"RETURNING ",
			strings.Join(options.Returning, ", "),
		},
		"",
	)
}

type selectParams struct {
	From    string
	Where   []WhereClause
	OrderBy []orderBy
	Offset  bool
	Limit   bool
	Columns []string
}

type orderBy struct {
	Expression string
	Direction  string
}

func buildSelectQuery(options selectParams) string {
	sql := strings.Join([]string{"SELECT ", strings.Join(options.Columns, ", "), " FROM ", options.From}, "")

	var placeholderOffset uint = 0
	if len(options.Where) > 0 {
		whereSql, replacements := buildWhereClause(options.Where)
		sql = strings.Join([]string{sql, " WHERE ", whereSql}, "")

		placeholderOffset += replacements
	}

	if len(options.OrderBy) > 0 {
		orderByClauses := make([]string, len(options.OrderBy))

		for idx, o := range options.OrderBy {
			orderByClauses[idx] = strings.Join([]string{o.Expression, o.Direction}, " ")
		}

		sql = strings.Join([]string{sql, " ORDER BY ", strings.Join(orderByClauses, ", ")}, "")
	}

	if options.Limit {
		sql = strings.Join([]string{sql, " LIMIT ", fmt.Sprintf("$%d", placeholderOffset+1)}, "")
		placeholderOffset++
	}

	if options.Offset {
		sql = strings.Join([]string{sql, " OFFSET ", fmt.Sprintf("$%d", placeholderOffset+1)}, "")
		placeholderOffset++
	}

	return sql
}

func buildWhereClause(clauses []WhereClause) (string, uint) {
	if len(clauses) == 0 {
		return "", 0
	}

	sql := joinWhereClauses(clauses)

	return replacePlaceholder(sql, 0)
}

func joinWhereClauses(clauses []WhereClause) string {
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

func columnEq(column string) string {
	return fmt.Sprintf("%s = ?", column)
}

func parens(expr string) string {
	return strings.Join([]string{"(", expr, ")"}, "")
}

func replacePlaceholder(expr string, offset uint) (string, uint) {
	matches := indexMatches(expr)

	sql := expr

	// Reverse the slice so the indexes are still valid as we modify the SQL.
	slices.Reverse(matches)

	for idx, match := range matches {
		left := sql[:match]
		right := sql[match+1:]

		// This is a little more complicated because we're iterating in reverse.
		placeholder := (len(matches) - idx) + int(offset)

		sql = strings.Join([]string{left, fmt.Sprintf("$%d", placeholder), right}, "")
	}

	return sql, uint(len(matches))
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

		// Pretty sure most SQL implementations escape quotes by doubling up
		// escaping = expr[i] == '\\'
	}

	return matches
}
