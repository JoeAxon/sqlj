package sqlj

import "testing"

func TestBuildWhereClause(t *testing.T) {
	result, replacements := buildWhereClause([]WhereClause{})

	if result != "" || replacements != 0 {
		t.Fatalf("Expected an empty string, got: %s\n", result)
	}

	result, replacements = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"id = ?"}},
	})

	if result != "id = $1" {
		t.Log("\"", result, "\"")
		t.Fatalf("Simple where clause failed")
	}

	if replacements != 1 {
		t.Fatalf("Expected 1 replacement, got: %d\n", replacements)
	}

	result, replacements = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"post_type = ?"}},
		{"AND", SimpleExpr{"created_at > ?"}},
	})

	if result != "post_type = $1 AND created_at > $2" {
		t.Fatalf("Multiple AND where clause failed: %s\n", result)
	}

	if replacements != 2 {
		t.Fatalf("Expected 2 replacements, got: %d\n", replacements)
	}

	result, replacements = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"post_type = ?"}},
		{"OR", SimpleExpr{"title = ?"}},
	})

	if result != "post_type = $1 OR title = $2" {
		t.Fatalf("Multiple AND OR where clause failed: %s\n", result)
	}

	if replacements != 2 {
		t.Fatalf("Expected 2 replacements, got: %d\n", replacements)
	}

	result, replacements = buildWhereClause([]WhereClause{
		{"AND", SimpleExpr{"id = ?"}},
		{"AND", NestedExpr{[]WhereClause{
			{"AND", SimpleExpr{"post_type = ?"}},
			{"OR", SimpleExpr{"title = ?"}},
		}}},
	})

	if result != "id = $1 AND (post_type = $2 OR title = $3)" {
		t.Fatalf("Nested expression failed: %s\n", result)
	}

	if replacements != 3 {
		t.Fatalf("Expected 3 replacements, got: %d\n", replacements)
	}
}

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

func TestReplacePlaceholder(t *testing.T) {
	result, replacements := replacePlaceholder("a = b", 0)

	if replacements != 0 && result != "a = b" {
		t.Fatalf("Expected no replacements, got: %d - %s", replacements, result)
	}

	result, replacements = replacePlaceholder("a = ?", 0)

	if replacements != 1 && result != "a = $1" {
		t.Fatalf("Expected 1 replacement, got: %d - %s", replacements, result)
	}

	result, replacements = replacePlaceholder("name = ?", 3)

	if replacements != 1 && result != "name = $4" {
		t.Fatalf("Expected 1 replacement, got: %d - %s", replacements, result)
	}
}
