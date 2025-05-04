package loaders

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type StreamLoader struct {
	Fields []string
}

func (s *StreamLoader) Load(ctx context.Context, client *goredis.Client, key string, rows []map[string]any) error {
	if len(s.Fields) == 0 {
		return fmt.Errorf("stream loader: no fields specified")
	}

	for _, row := range rows {
		values := make(map[string]interface{})

		for _, field := range s.Fields {
			val, ok := row[field]
			if !ok {
				return fmt.Errorf("stream loader: field '%s' not found in row", field)
			}
			values[field] = val
		}

		err := client.XAdd(ctx, &goredis.XAddArgs{
			Stream: key,
			Values: values,
		}).Err()
		if err != nil {
			return fmt.Errorf("stream loader: XADD failed: %w", err)
		}
	}

	return nil
}
