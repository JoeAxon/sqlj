package sqlj

import (
	"fmt"
	"strings"
)

func buildInsertSQL(table string, fields []Field, columns []string) string {
	names := make([]string, len(fields))
	placeholders := make([]string, len(fields))

	n := 0
	for idx, f := range fields {
		names[idx] = f.GetName()
		placeholders[idx] = f.GetPlaceholder(n)

		if !f.IsLiteral() {
			n++
		}
	}

	return strings.Join(
		[]string{
			"INSERT INTO ",
			table,
			" (",
			strings.Join(names, ", "),
			") VALUES (",
			strings.Join(placeholders, ", "),
			") RETURNING ",
			strings.Join(columns, ", "),
		},
		"",
	)
}

func buildUpdateSQL(table string, fields []Field, columns []string) string {
	setExpressions := make([]string, len(fields))
	for idx, f := range fields {
		setExpressions[idx] = fmt.Sprintf("%s = %s", f.GetName(), f.GetPlaceholder(idx))
	}

	return strings.Join(
		[]string{
			"UPDATE ",
			table,
			" SET ",
			strings.Join(setExpressions, ", "),
			fmt.Sprintf(" WHERE id = $%d ", len(fields)),
			"RETURNING ",
			strings.Join(columns, ", "),
		},
		"",
	)
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
