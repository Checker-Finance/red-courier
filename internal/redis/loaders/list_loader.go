package loaders

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type ListLoader struct {
	ValueField string
}

func (l *ListLoader) Load(ctx context.Context, client *goredis.Client, key string, rows []map[string]any) error {
	var values []interface{}

	for _, row := range rows {
		val, ok := row[l.ValueField]
		if !ok {
			return fmt.Errorf("list loader: field '%s' not found in row", l.ValueField)
		}
		values = append(values, fmt.Sprintf("%v", val))
	}

	if len(values) == 0 {
		return nil
	}

	return client.LPush(ctx, key, values...).Err()
}
