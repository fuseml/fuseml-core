package util

// StringInSlice verifies if a string slice contains a string value
func StringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

// RefString converts a string value into a string reference. The reference can also take
// a nil value to indicate a default value
func RefString(s string, defaultValue ...string) *string {
	ds := ""
	if len(defaultValue) > 0 {
		ds = defaultValue[0]
	}
	if s == ds {
		return nil
	}
	return &s
}

// DerefString converts a string reference into a string value. If the reference is nil,
// the default value is returned instead
func DerefString(s *string, defaultValue ...string) string {
	ds := ""
	if len(defaultValue) > 0 {
		ds = defaultValue[0]
	}
	if s != nil {
		return *s
	}
	return ds
}

// RefBool converts a bool value into a bool reference. The reference can also take
// a nil value to indicate a default value
func RefBool(b bool, defaultValue ...bool) *bool {
	db := false
	if len(defaultValue) > 0 {
		db = defaultValue[0]
	}
	if b == db {
		return nil
	}
	return &b
}

// DerefBool converts a bool reference into a bool value. If the reference is nil,
// the default value is returned instead
func DerefBool(s *bool, defaultValue ...bool) bool {
	db := false
	if len(defaultValue) > 0 {
		db = defaultValue[0]
	}
	if s != nil {
		return *s
	}
	return db
}
