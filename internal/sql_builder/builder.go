package sqlbuilder

import (
	"fmt"
	"path"
	"strings"
)

type TrackingSpec struct {
	Column       string // resolved db column name
	Operator     string // e.g. ">" or ">="
	LastValueKey string
}

// Input for building a SELECT query.
type SelectSpec struct {
	Schema    string   // defaults to "public" if empty
	Table     string   // REQUIRED (bare table name; no schema)
	Columns   []string // REQUIRED
	Where     string   // optional raw sql (without "WHERE")
	Tracking  *TrackingSpec
	LastValue *string // optional; if nil/"" => first run
}

// Output plan for DB execution.
type SelectPlan struct {
	SQL         string
	Args        []any
	FirstRun    bool   // true when no last checkpoint (i.e., we didn't add tracking predicate)
	TrackingCol string // resolved tracking column if present
}

// Parse "schema.table" or "table" into (schema, table).
func SplitSchemaTable(qualified string) (string, string) {
	if strings.Contains(qualified, ".") {
		// be robust to multiple dots; we only split at first
		parts := strings.SplitN(qualified, ".", 2)
		return parts[0], parts[1]
	}
	return "public", qualified
}

// Build a SELECT plan with optional static WHERE and optional tracking clause.
// - If LastValue is nil/empty and Tracking != nil => FirstRun=true and no tracking predicate is appended.
// - If LastValue is present => append tracking predicate with positional arg $N.
func BuildSelect(spec SelectSpec) (SelectPlan, error) {
	if len(spec.Columns) == 0 {
		return SelectPlan{}, fmt.Errorf("no columns provided")
	}
	if spec.Table == "" {
		return SelectPlan{}, fmt.Errorf("no table provided")
	}
	schema := spec.Schema
	if schema == "" {
		schema = "public"
	}

	quotedSchema := `"` + strings.ReplaceAll(schema, `"`, `""`) + `"`
	quotedTable := `"` + strings.ReplaceAll(spec.Table, `"`, `""`) + `"`

	sql := fmt.Sprintf(`SELECT %s FROM %s.%s`,
		strings.Join(spec.Columns, ", "),
		quotedSchema, quotedTable,
	)

	var clauses []string
	var args []any
	firstRun := false
	var trackingCol string

	// static WHERE
	if spec.Where != "" {
		clauses = append(clauses, spec.Where)
	}

	// tracking
	if spec.Tracking != nil {
		trackingCol = spec.Tracking.Column
		if trackingCol == "" {
			return SelectPlan{}, fmt.Errorf("tracking column is empty")
		}
		if spec.LastValue == nil || *spec.LastValue == "" {
			// No checkpoint -> FIRST RUN -> do not add tracking predicate
			firstRun = true
		} else {
			// Add param with correct index
			paramIdx := len(args) + 1
			clauses = append(clauses, fmt.Sprintf("%s %s $%d", trackingCol, spec.Tracking.Operator, paramIdx))
			args = append(args, *spec.LastValue)
		}
	}

	if len(clauses) > 0 {
		sql += " WHERE " + strings.Join(clauses, " AND ")
	}

	return SelectPlan{
		SQL:         sql,
		Args:        args,
		FirstRun:    firstRun,
		TrackingCol: trackingCol,
	}, nil
}

// Helper to resolve schema/table + columns from a "table" value that may include a schema.
// Returns (SelectSpec, bareTableName).
func FromQualifiedTable(qualified string, columns []string, where string, tracking *TrackingSpec, lastVal *string) (SelectSpec, string) {
	schema, tbl := SplitSchemaTable(qualified)
	return SelectSpec{
		Schema:    schema,
		Table:     tbl,
		Columns:   columns,
		Where:     where,
		Tracking:  tracking,
		LastValue: lastVal,
	}, tbl
}

// Optional convenience to build a safe alias like schema_table for stream names, etc.
func SafeAlias(schema, table string) string {
	return strings.ReplaceAll(path.Join(schema, table), "/", "_")
}
