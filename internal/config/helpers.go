package config

func (t TaskConfig) EffectiveLogSQL(appDefault bool) bool {
	if t.LogSQL != nil {
		return *t.LogSQL
	}
	return appDefault
}
