// internal/sqlbuilder/builder_test.go
package sqlbuilder

import (
	"reflect"
	"testing"
)

func TestBuild_NoWhere_NoTracking(t *testing.T) {
	spec := SelectSpec{
		Schema:  "public",
		Table:   "orders",
		Columns: []string{"id", "status"},
	}
	plan, err := BuildSelect(spec)
	if err != nil {
		t.Fatal(err)
	}
	wantSQL := `SELECT id, status FROM "public"."orders"`
	if plan.SQL != wantSQL {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", plan.SQL, wantSQL)
	}
	if len(plan.Args) != 0 || plan.FirstRun != false {
		t.Fatalf("unexpected args/firstRun: %+v", plan)
	}
}

func TestBuild_WithWhere_Only(t *testing.T) {
	spec := SelectSpec{
		Schema:  "public",
		Table:   "orders",
		Columns: []string{"id"},
		Where:   "status = 'NEW'",
	}
	plan, err := BuildSelect(spec)
	if err != nil {
		t.Fatal(err)
	}
	want := `SELECT id FROM "public"."orders" WHERE status = 'NEW'`
	if plan.SQL != want {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", plan.SQL, want)
	}
}

func TestBuild_WithTracking_FirstRun(t *testing.T) {
	spec := SelectSpec{
		Schema:  "public",
		Table:   "orders",
		Columns: []string{"id", "created_at"},
		Tracking: &TrackingSpec{
			Column:   "created_at",
			Operator: ">",
		},
		// LastValue nil => first run
	}
	plan, err := BuildSelect(spec)
	if err != nil {
		t.Fatal(err)
	}
	want := `SELECT id, created_at FROM "public"."orders"`
	if plan.SQL != want || !plan.FirstRun {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuild_WithWhere_And_Tracking(t *testing.T) {
	last := "2025-09-18T00:00:00Z"
	spec := SelectSpec{
		Schema:  "public",
		Table:   "orders",
		Columns: []string{"id", "amount", "created_at"},
		Where:   "amount > 1000",
		Tracking: &TrackingSpec{
			Column:   "created_at",
			Operator: ">",
		},
		LastValue: &last,
	}
	plan, err := BuildSelect(spec)
	if err != nil {
		t.Fatal(err)
	}
	wantSQL := `SELECT id, amount, created_at FROM "public"."orders" WHERE amount > 1000 AND created_at > $1`
	if plan.SQL != wantSQL {
		t.Fatalf("sql mismatch:\n got: %s\nwant: %s", plan.SQL, wantSQL)
	}
	if !reflect.DeepEqual(plan.Args, []any{last}) {
		t.Fatalf("args mismatch: %+v", plan.Args)
	}
	if plan.FirstRun {
		t.Fatalf("expected FirstRun=false")
	}
}
