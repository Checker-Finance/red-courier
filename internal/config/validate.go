// internal/config/validate.go
package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
)

var validate = validator.New()

func Validate(cfg *Config) error {
	if err := validate.Struct(cfg); err != nil {
		return err
	}
	// Per-task custom rules
	for _, t := range cfg.Tasks {
		// schedule: allow robfig cron or "@every"
		if err := validateSchedule(t.Schedule); err != nil {
			return fmt.Errorf("task %q: %w", t.Name, err)
		}
		// tracking column must be among resolved fields (after alias resolution if any)
		if t.Tracking != nil {
			if !contains(t.Fields, t.Tracking.Column) {
				return fmt.Errorf("task %q: tracking.column %q not in fields", t.Name, t.Tracking.Column)
			}
		}
		// table must be "schema.table" or bare "table"
		if strings.Count(t.Table, ".") > 1 {
			return fmt.Errorf("task %q: invalid table %q", t.Name, t.Table)
		}
	}
	return nil
}

func validateSchedule(s string) error {
	if strings.HasPrefix(s, "@every ") {
		return nil
	}
	_, err := cron.ParseStandard(s)
	if err != nil {
		return fmt.Errorf("invalid schedule: %v", err)
	}
	return nil
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
