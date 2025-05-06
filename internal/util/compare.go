package util

import (
	"fmt"
	"time"
)

func CompareAny(a, b any) int {
	switch aTyped := a.(type) {
	case time.Time:
		bTyped, ok := b.(time.Time)
		if !ok {
			return 0
		}
		if aTyped.After(bTyped) {
			return 1
		}
		if aTyped.Before(bTyped) {
			return -1
		}
		return 0
	case int, int64, float64:
		af := toFloat64(a)
		bf := toFloat64(b)
		if af > bf {
			return 1
		}
		if af < bf {
			return -1
		}
		return 0
	case string:
		bTyped, ok := b.(string)
		if !ok {
			return 0
		}
		if aTyped > bTyped {
			return 1
		}
		if aTyped < bTyped {
			return -1
		}
		return 0
	default:
		return 0
	}
}

func toFloat64(x any) float64 {
	switch v := x.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

func ToRedisString(val any) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	case time.Time:
		return v.Format(time.RFC3339Nano), true
	case int, int64, float64:
		return fmt.Sprintf("%v", v), true
	default:
		return "", false
	}
}
