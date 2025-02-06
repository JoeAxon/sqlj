package sqlj

import (
	"fmt"
	"strings"
)

type Delete struct {
	From  string
	Where []WhereClause
}

func buildDeleteSQL(options Delete) string {
	sql := strings.Join([]string{"DELETE FROM ", options.From}, "")

	if len(options.Where) > 0 {
		whereSQL, _ := buildWhereClause(options.Where)

		sql = strings.Join([]string{sql, " WHERE ", whereSQL}, "")
	}

	return sql
}

type Insert struct {
	From      string
	Fields    []Field
	Returning []string
}

func buildInsertSQL(options Insert) string {
	names := make([]string, len(options.Fields))
	placeholders := make([]string, len(options.Fields))

	n := 0
	for idx, f := range options.Fields {
		names[idx] = f.GetName()
		placeholders[idx] = f.GetPlaceholder(n)

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

type Update struct {
	From      string
	Fields    []Field
	Returning []string
}

func buildUpdateSQL(options Update) string {
	setExpressions := make([]string, len(options.Fields))
	for idx, f := range options.Fields {
		setExpressions[idx] = fmt.Sprintf("%s = %s", f.GetName(), f.GetPlaceholder(idx))
	}

	return strings.Join(
		[]string{
			"UPDATE ",
			options.From,
			" SET ",
			strings.Join(setExpressions, ", "),
			fmt.Sprintf(" WHERE id = $%d ", len(options.Fields)),
			"RETURNING ",
			strings.Join(options.Returning, ", "),
		},
		"",
	)
}

type Select struct {
	From    string
	Where   []WhereClause
	OrderBy []OrderBy
	Offset  bool
	Limit   bool
	Columns []string
}

func buildSelectQuery(options Select) string {
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
			orderByClauses[idx] = strings.Join([]string{o.expression, o.direction}, " ")
		}

		sql = strings.Join([]string{sql, " ORDER BY ", strings.Join(orderByClauses, ", ")}, "")
	}

	if options.Offset {
		sql = strings.Join([]string{sql, " OFFSET ", fmt.Sprintf("$%d", placeholderOffset)}, "")
		placeholderOffset++
	}

	if options.Limit {
		sql = strings.Join([]string{sql, " LIMIT ", fmt.Sprintf("$%d", placeholderOffset)}, "")
		placeholderOffset++
	}

	return sql
}

func buildWhereClause(clauses []WhereClause) (string, uint) {
	if len(clauses) == 0 {
		return "", 0
	}

	sql := make([]string, len(clauses)*2)

	var n uint = 0
	for idx, clause := range clauses {
		if idx == 0 {
			sql[idx*2] = ""
		} else {
			sql[idx*2] = clause.Type
		}

		expr, replacementCount := replacePlaceholder(clause.Expr.String(), n)

		sql[idx*2+1] = expr

		n += replacementCount
	}

	return strings.Join(sql[1:], " "), n
}

func columnEq(column string) string {
	return fmt.Sprintf("%s = ?", column)
}

func parens(expr string) string {
	return strings.Join([]string{"(", expr, ")"}, "")
}

func replacePlaceholder(expr string, offset uint) (string, uint) {
	matches := indexMatches(expr)

	pieces := make([]string, len(matches)*3)

	for idx, match := range matches {
		left := expr[:match]
		right := expr[match+1:]

		pieces[idx] = left
		pieces[idx+1] = fmt.Sprintf("$%d", idx+int(offset))
		pieces[idx+2] = right
	}

	return strings.Join(pieces, ""), 0
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
