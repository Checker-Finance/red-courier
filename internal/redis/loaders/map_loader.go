package loaders

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type MapLoader struct {
	KeyField   string
	ValueField string
}

func (m *MapLoader) Load(ctx context.Context, client *goredis.Client, key string, rows []map[string]any) error {
	data := make(map[string]string)

	for _, row := range rows {
		k, ok1 := row[m.KeyField]
		v, ok2 := row[m.ValueField]

		if !ok1 || !ok2 {
			return fmt.Errorf("map loader: missing field '%s' or '%s'", m.KeyField, m.ValueField)
		}

		data[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
	}

	if len(data) == 0 {
		return nil
	}

	return client.HSet(ctx, key, data).Err()
}
