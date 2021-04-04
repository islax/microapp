package value

import "time"

// GetBoolPointer gets pointer to the given value
func GetBoolPointer(val bool) *bool {
	return &val
}

// GetIntPointer gets pointer to the given value
func GetIntPointer(val int) *int {
	return &val
}

// GetInt32Pointer gets pointer to the given value
func GetInt32Pointer(val int32) *int32 {
	return &val
}

// GetInt64Pointer gets pointer to the given value
func GetInt64Pointer(val int64) *int64 {
	return &val
}

// GetStringPointer gets pointer to the given value
func GetStringPointer(val string) *string {
	return &val
}

// GetTimePointer gets pointer to the given value
func GetTimePointer(val time.Time) *time.Time {
	return &val
}
