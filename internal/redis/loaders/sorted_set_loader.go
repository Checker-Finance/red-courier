package loaders

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"
)

type SortedSetLoader struct {
	ValueField string
	ScoreField string
}

func (s *SortedSetLoader) Load(ctx context.Context, client *goredis.Client, key string, rows []map[string]any) error {
	var members []goredis.Z

	for _, row := range rows {
		valRaw, valOk := row[s.ValueField]
		scoreRaw, scoreOk := row[s.ScoreField]

		if !valOk || !scoreOk {
			return fmt.Errorf("sorted_set loader: missing required fields '%s' or '%s'", s.ValueField, s.ScoreField)
		}

		scoreStr := fmt.Sprintf("%v", scoreRaw)
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return fmt.Errorf("sorted_set loader: invalid score '%s': %w", scoreStr, err)
		}

		val := fmt.Sprintf("%v", valRaw)
		members = append(members, goredis.Z{
			Score:  score,
			Member: val,
		})
	}

	if len(members) == 0 {
		return nil
	}

	return client.ZAdd(ctx, key, members...).Err()
}
