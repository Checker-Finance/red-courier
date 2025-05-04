package loaders

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type SetLoader struct {
	ValueField string
}

func (s *SetLoader) Load(ctx context.Context, client *goredis.Client, key string, rows []map[string]any) error {
	var values []interface{}

	for _, row := range rows {
		val, ok := row[s.ValueField]
		if !ok {
			return fmt.Errorf("set loader: field '%s' not found in row", s.ValueField)
		}
		values = append(values, fmt.Sprintf("%v", val))
	}

	if len(values) == 0 {
		return nil
	}

	return client.SAdd(ctx, key, values...).Err()
}
