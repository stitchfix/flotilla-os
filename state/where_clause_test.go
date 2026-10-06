package state

import (
	"testing"

	"github.com/stitchfix/flotilla-os/exceptions"
)

// newClauseTestManager returns a bare SQLStateManager suitable for exercising
// the where-clause builders, which are pure string helpers and never touch the
// database handle.
func newClauseTestManager() *SQLStateManager {
	return &SQLStateManager{}
}

func TestMakeWhereClause_SingleValueIsEscaped(t *testing.T) {
	sm := newClauseTestManager()

	// "engine" is a plain equality field (not a like/_since/_until field).
	wc, err := sm.makeWhereClause(map[string][]string{"engine": {"eks"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wc) != 1 {
		t.Fatalf("expected 1 clause, got %d: %v", len(wc), wc)
	}
	if wc[0] != "engine='eks'" {
		t.Errorf("expected engine='eks', got %q", wc[0])
	}
}

// This is the DP-5721 regression: a value that already contains single quotes
// must be escaped into a valid literal rather than breaking out of the string
// and leaving a bare token (which produced `syntax error at or near "eks"`).
func TestMakeWhereClause_QuotedValueDoesNotBreakOut(t *testing.T) {
	sm := newClauseTestManager()

	wc, err := sm.makeWhereClause(map[string][]string{"engine": {"'eks'"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// pq.QuoteLiteral doubles the embedded quotes: '''eks'''
	if wc[0] != "engine='''eks'''" {
		t.Errorf("expected escaped literal engine='''eks''', got %q", wc[0])
	}
}

func TestMakeWhereClause_InjectionAttemptInValueIsInert(t *testing.T) {
	sm := newClauseTestManager()

	wc, err := sm.makeWhereClause(map[string][]string{"engine": {"x' OR '1'='1"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc[0] != "engine='x'' OR ''1''=''1'" {
		t.Errorf("expected injection to be escaped into a literal, got %q", wc[0])
	}
}

func TestMakeWhereClause_MultipleValuesEscapedInInClause(t *testing.T) {
	sm := newClauseTestManager()

	wc, err := sm.makeWhereClause(map[string][]string{"engine": {"a'b", "c"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc[0] != "engine in ('a''b','c')" {
		t.Errorf("expected escaped IN clause, got %q", wc[0])
	}
}

func TestMakeWhereClause_LikeFieldEscaped(t *testing.T) {
	sm := newClauseTestManager()

	// "image" is a like field.
	wc, err := sm.makeWhereClause(map[string][]string{"image": {"my'img"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc[0] != "image like '%my''img%'" {
		t.Errorf("expected escaped like clause, got %q", wc[0])
	}
}

func TestMakeWhereClause_SinceFieldFormatsComparison(t *testing.T) {
	sm := newClauseTestManager()

	wc, err := sm.makeWhereClause(map[string][]string{"started_at_since": {"2020-01-01"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc[0] != "started_at > '2020-01-01'" {
		t.Errorf("expected started_at > '2020-01-01', got %q", wc[0])
	}
}

func TestMakeWhereClause_UntilFieldFormatsComparison(t *testing.T) {
	sm := newClauseTestManager()

	wc, err := sm.makeWhereClause(map[string][]string{"started_at_until": {"2020-01-01"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc[0] != "started_at < '2020-01-01'" {
		t.Errorf("expected started_at < '2020-01-01', got %q", wc[0])
	}
}

func TestMakeWhereClause_InvalidFieldNameRejected(t *testing.T) {
	sm := newClauseTestManager()

	badKeys := []string{
		"1=1) UNION SELECT", // injection via field name
		"engine'",           // stray quote
		"group name",        // whitespace
		"drop;table",        // semicolon
	}
	for _, k := range badKeys {
		_, err := sm.makeWhereClause(map[string][]string{k: {"x"}})
		if err == nil {
			t.Errorf("expected error for invalid field name %q, got nil", k)
			continue
		}
		if _, ok := err.(exceptions.MalformedInput); !ok {
			t.Errorf("expected MalformedInput for field name %q, got %T", k, err)
		}
	}
}

func TestMakeEnvWhereClause_EscapesValues(t *testing.T) {
	sm := newClauseTestManager()

	wc := sm.makeEnvWhereClause(map[string]string{"FOO": "ba'r"})
	if len(wc) != 1 {
		t.Fatalf("expected 1 clause, got %d: %v", len(wc), wc)
	}
	// The JSON literal must be a single, properly escaped SQL string literal;
	// the embedded single quote is doubled by pq.QuoteLiteral.
	if wc[0] != `env @> '[{"name":"FOO","value":"ba''r"}]'` {
		t.Errorf("expected escaped env clause, got %q", wc[0])
	}
}
