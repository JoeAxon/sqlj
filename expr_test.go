package sqlj

import "testing"

func TestBuildWhereClause(t *testing.T) {
	result := buildWhereClause([]WhereClause{})

	if result != "" {
		t.Fatalf("Expected an empty string, got: %s\n", result)
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"id = ?"}},
	})

	if result != "id = ?" {
		t.Log("\"", result, "\"")
		t.Fatalf("Simple where clause failed")
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"post_type = ?"}},
		{"AND", SimpleExpr{"created_at > ?"}},
	})

	if result != "post_type = ? AND created_at > ?" {
		t.Fatalf("Multiple AND where clause failed")
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"post_type = ?"}},
		{"OR", SimpleExpr{"title = ?"}},
	})

	if result != "post_type = ? OR title = ?" {
		t.Fatalf("Multiple AND OR where clause failed")
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"id = ?"}},
		{"AND", NestedExpr{[]WhereClause{
			{"AND", SimpleExpr{"post_type = ?"}},
			{"OR", SimpleExpr{"title = ?"}},
		}}},
	})

	if result != "id = ? AND (post_type = ? OR title = ?)" {
		t.Fatalf("Nested expression failed")
	}
}
