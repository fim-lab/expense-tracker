package domain

import "testing"

func TestParseSearchQuery_Matches(t *testing.T) {
	t.Run("plain query is a case-insensitive substring match", func(t *testing.T) {
		q := ParseSearchQuery("holiday")
		if !q.Matches("Holiday Hotel") {
			t.Errorf("expected match")
		}
		if !q.Matches("a HOLIDAY booking") {
			t.Errorf("expected case-insensitive match")
		}
		if q.Matches("Bike Chain") {
			t.Errorf("expected no match")
		}
	})

	t.Run("AND requires both terms", func(t *testing.T) {
		q := ParseSearchQuery("bike AND holiday")
		if !q.Matches("Holiday with a bike") {
			t.Errorf("expected match on 'Holiday with a bike'")
		}
		if q.Matches("Holiday Hotel") {
			t.Errorf("expected no match on 'Holiday Hotel'")
		}
	})

	t.Run("AND NOT excludes the second term", func(t *testing.T) {
		q := ParseSearchQuery("bike AND NOT holiday")
		if !q.Matches("Bike Chain") {
			t.Errorf("expected match on 'Bike Chain'")
		}
		if q.Matches("Holiday with a bike") {
			t.Errorf("expected no match on 'Holiday with a bike'")
		}
	})

	t.Run("OR matches either term", func(t *testing.T) {
		q := ParseSearchQuery("bike OR holiday")
		if !q.Matches("Bike Chain") {
			t.Errorf("expected match on 'Bike Chain'")
		}
		if !q.Matches("Holiday Hotel") {
			t.Errorf("expected match on 'Holiday Hotel'")
		}
		if q.Matches("Groceries") {
			t.Errorf("expected no match on 'Groceries'")
		}
	})

	t.Run("leading NOT negates a single term", func(t *testing.T) {
		q := ParseSearchQuery("NOT holiday")
		if !q.Matches("Bike Chain") {
			t.Errorf("expected match on 'Bike Chain'")
		}
		if q.Matches("Holiday Hotel") {
			t.Errorf("expected no match on 'Holiday Hotel'")
		}
	})

	t.Run("keywords are case sensitive, lowercase stays part of the term", func(t *testing.T) {
		q := ParseSearchQuery("bike and holiday")
		if !q.Matches("bike and holiday supplies") {
			t.Errorf("expected 'and' to be treated as literal text")
		}
		if q.Matches("Holiday with a bike") {
			t.Errorf("did not expect a match, 'and' should not act as a keyword")
		}
	})

	t.Run("multi-word terms are preserved between keywords", func(t *testing.T) {
		q := ParseSearchQuery("bike shop AND NOT holiday sale")
		if !q.Matches("Local Bike Shop") {
			t.Errorf("expected match on 'Local Bike Shop'")
		}
		if q.Matches("Bike Shop Holiday Sale") {
			t.Errorf("expected no match on 'Bike Shop Holiday Sale'")
		}
	})

	t.Run("NOT binds tighter than AND, which binds tighter than OR", func(t *testing.T) {
		q := ParseSearchQuery("bike AND holiday OR NOT cake")
		if !q.Matches("Holiday with a bike") {
			t.Errorf("expected match via 'bike AND holiday'")
		}
		if !q.Matches("Groceries") {
			t.Errorf("expected match via 'NOT cake'")
		}
		if q.Matches("Chocolate Cake") {
			t.Errorf("expected no match: fails both 'bike AND holiday' and 'NOT cake'")
		}
	})
}

func TestParseSearchQuery_SQLCondition(t *testing.T) {
	t.Run("plain term renders a single ILIKE", func(t *testing.T) {
		var args []interface{}
		argID := 2
		cond := ParseSearchQuery("bike").SQLCondition("t.description", &args, &argID)
		if cond != "t.description ILIKE $2" {
			t.Errorf("unexpected condition: %s", cond)
		}
		if len(args) != 1 || args[0] != "%bike%" {
			t.Errorf("unexpected args: %v", args)
		}
		if argID != 3 {
			t.Errorf("expected argID to advance to 3, got %d", argID)
		}
	})

	t.Run("AND NOT renders a negated group with its own placeholder", func(t *testing.T) {
		var args []interface{}
		argID := 3
		cond := ParseSearchQuery("bike AND NOT holiday").SQLCondition("t.description", &args, &argID)
		if cond != "(t.description ILIKE $3 AND NOT (t.description ILIKE $4))" {
			t.Errorf("unexpected condition: %s", cond)
		}
		if len(args) != 2 || args[0] != "%bike%" || args[1] != "%holiday%" {
			t.Errorf("unexpected args: %v", args)
		}
		if argID != 5 {
			t.Errorf("expected argID to advance to 5, got %d", argID)
		}
	})
}
