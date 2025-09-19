// internal/config/loader_test.go
package config

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempYAML(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return path
}

//go:embed testdata/configs/*.yaml
var testConfigs embed.FS

func TestAllConfigFilesPassOrFailAsExpected(t *testing.T) {
	entries, err := fs.ReadDir(testConfigs, "testdata/configs")
	if err != nil {
		t.Fatalf("embed ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("no embedded test configs found")
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !(strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")) {
			continue
		}

		t.Run(name, func(t *testing.T) {
			b, err := fs.ReadFile(testConfigs, filepath.Join("testdata/configs", name))
			if err != nil {
				t.Fatalf("embed ReadFile %s: %v", name, err)
			}

			// write to a temp file to exercise the real loader path
			tmp := t.TempDir()
			p := filepath.Join(tmp, name)
			if err := os.WriteFile(p, b, 0o600); err != nil {
				t.Fatalf("write temp: %v", err)
			}

			cfg, loadErr := LoadConfig(p)
			isInvalid := strings.HasPrefix(name, "invalid_")

			if loadErr != nil {
				if isInvalid {
					t.Logf("expected load error for invalid %q: %v", name, loadErr)
					return
				}
				t.Fatalf("unexpected load error for %q: %v", name, loadErr)
			}

			valErr := Validate(cfg)
			if isInvalid && valErr == nil {
				t.Fatalf("expected validation to FAIL for %q, but it passed", name)
			}
			if !isInvalid && valErr != nil {
				t.Fatalf("expected validation to PASS for %q, got: %v", name, valErr)
			}
		})
	}
}

func TestLoadConfig_ThenValidate_OK(t *testing.T) {
	yaml := `
tasks:
  - name: quotes_stream
    table: public.quotes
    fields: [id, instrument, px, updated_at]
    schedule: "@every 10s"
    tracking:
      column: updated_at
      operator: ">="
      last_value_key: checkpoint:quotes_stream
`
	path := writeTempYAML(t, yaml)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate error: %v", err)
	}
}

func TestLoadConfig_AppliesDefaults_NoSchema_NoStructure(t *testing.T) {
	yaml := `
tasks:
  - name: orders_stream
    table: orders              # no schema provided
    fields: [id, status, created_at]
    schedule: "@every 30s"     # structure omitted -> defaults to "stream"
`
	path := writeTempYAML(t, yaml)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if len(cfg.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(cfg.Tasks))
	}

	task := cfg.Tasks[0]

	// default schema should be applied
	if got, want := task.Table, "public.orders"; got != want {
		t.Errorf("table mismatch: got %q want %q", got, want)
	}

	// default structure should be applied
	if got, want := task.Structure, "stream"; got != want {
		t.Errorf("structure mismatch: got %q want %q", got, want)
	}
}

func TestLoadConfig_PreservesProvidedSchema_AndStructure(t *testing.T) {
	yaml := `
tasks:
  - name: trades_snapshot
    table: analytics.trades     # schema provided -> should be preserved
    structure: snapshot         # explicit -> should be preserved
    fields: [trade_id, executed_at]
    schedule: "0 * * * *"
`
	path := writeTempYAML(t, yaml)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	task := cfg.Tasks[0]

	if got, want := task.Table, "analytics.trades"; got != want {
		t.Errorf("table mismatch: got %q want %q", got, want)
	}
	if got, want := task.Structure, "snapshot"; got != want {
		t.Errorf("structure mismatch: got %q want %q", got, want)
	}
}

func TestLoadConfig_InvalidYAML_ReturnsError(t *testing.T) {
	// deliberately malformed YAML
	yaml := `
tasks:
  - name: bad
    table: public.orders
    fields: [id, status
    schedule: "@every 10s"
`
	path := writeTempYAML(t, yaml)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatalf("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_WithWhereAndTracking_UnmarshalsFields(t *testing.T) {
	yaml := `
tasks:
  - name: high_value_new_orders
    table: public.orders
    fields: [id, status, amount, created_at]
    schedule: "@every 30s"
    where: "status = 'NEW' AND amount > 1000"
    tracking:
      column: created_at
      operator: ">"
      last_value_key: checkpoint:high_value_orders
`
	path := writeTempYAML(t, yaml)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	task := cfg.Tasks[0]
	if task.Where == "" {
		t.Errorf("expected where clause to be set")
	}
	if task.Tracking == nil {
		t.Fatalf("expected tracking to be set")
	}
	if got, want := task.Tracking.Column, "created_at"; got != want {
		t.Errorf("tracking.column mismatch: got %q want %q", got, want)
	}
}
