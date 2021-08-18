package svc

func refString(s string, defaultValue ...string) *string {
	ds := ""
	if len(defaultValue) > 0 {
		ds = defaultValue[0]
	}
	if s == ds {
		return nil
	}
	return &s
}

func derefString(s *string, defaultValue ...string) string {
	ds := ""
	if len(defaultValue) > 0 {
		ds = defaultValue[0]
	}
	if s != nil {
		return *s
	}
	return ds
}

func refBool(b bool, defaultValue ...bool) *bool {
	db := false
	if len(defaultValue) > 0 {
		db = defaultValue[0]
	}
	if b == db {
		return nil
	}
	return &b
}

func derefBool(s *bool, defaultValue ...bool) bool {
	db := false
	if len(defaultValue) > 0 {
		db = defaultValue[0]
	}
	if s != nil {
		return *s
	}
	return db
}
