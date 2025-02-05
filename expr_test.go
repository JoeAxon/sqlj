package sqlj

import "testing"

/*
func TestBuildWhereClause(t *testing.T) {
	result := buildWhereClause([]WhereClause{})

	if result != "" {
		t.Fatalf("Expected an empty string, got: %s\n", result)
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"id = ?"}},
	})

	if result != "id = $0" {
		t.Log("\"", result, "\"")
		t.Fatalf("Simple where clause failed")
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"post_type = ?"}},
		{"AND", SimpleExpr{"created_at > ?"}},
	})

	if result != "post_type = $0 AND created_at > $1" {
		t.Fatalf("Multiple AND where clause failed")
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"post_type = ?"}},
		{"OR", SimpleExpr{"title = ?"}},
	})

	if result != "post_type = $0 OR title = $1" {
		t.Fatalf("Multiple AND OR where clause failed")
	}

	result = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"id = ?"}},
		{"AND", NestedExpr{[]WhereClause{
			{"AND", SimpleExpr{"post_type = ?"}},
			{"OR", SimpleExpr{"title = ?"}},
		}}},
	})

	if result != "id = $0 AND (post_type = $1 OR title = $2)" {
		t.Fatalf("Nested expression failed")
	}
}
*/

func TestIndexMatches(t *testing.T) {
	result := indexMatches("something = nothing")

	if len(result) != 0 {
		t.Fatalf("Expected 0 matches, got: %d\n", len(result))
	}

	result = indexMatches("something = ?")

	if len(result) != 1 {
		t.Fatalf("Expected 1 match, got: %d\n", len(result))
	}

	if result[0] != 12 {
		t.Fatalf("Expected index 12, got: %d\n", result[0])
	}

	result = indexMatches("something = '?'")

	if len(result) != 0 {
		t.Fatalf("Expected 0 matches, got: %d\n", len(result))
	}

	result = indexMatches("something = '?' || ?")

	if len(result) != 1 {
		t.Fatalf("Expected 1 match, got: %d\n", len(result))
	}

	if result[0] != 19 {
		t.Fatalf("Expected index 19, got: %d\n", result[0])
	}
}
