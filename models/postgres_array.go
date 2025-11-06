package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// StringArray is a custom type for PostgreSQL text[] array
type StringArray []string

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	// Format: {"value1","value2"}
	values := make([]string, len(a))
	for i, v := range a {
		// Escape quotes in values
		v = strings.ReplaceAll(v, `"`, `\"`)
		values[i] = `"` + v + `"`
	}
	return "{" + strings.Join(values, ",") + "}", nil
}

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return errors.New("cannot scan non-string value into StringArray")
	}

	// Handle PostgreSQL array format: {value1,value2} or {"value1","value2"}
	if str == "" || str == "{}" {
		*a = StringArray{}
		return nil
	}

	// Remove curly braces
	str = strings.Trim(str, "{}")
	if str == "" {
		*a = StringArray{}
		return nil
	}

	// Parse PostgreSQL array format
	// Simple approach: split by comma, handle quoted strings
	var result []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(str); i++ {
		char := str[i]

		if char == '"' {
			if i+1 < len(str) && str[i+1] == '"' {
				// Escaped quote
				current.WriteByte('"')
				i++
				continue
			}
			inQuotes = !inQuotes
			continue
		}

		if char == ',' && !inQuotes {
			val := strings.TrimSpace(current.String())
			if val != "" {
				result = append(result, val)
			}
			current.Reset()
			continue
		}

		current.WriteByte(char)
	}

	// Add last value
	val := strings.TrimSpace(current.String())
	if val != "" {
		result = append(result, val)
	}

	*a = StringArray(result)
	return nil
}

// MarshalJSON implements json.Marshaler
func (a StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(a))
}

// UnmarshalJSON implements json.Unmarshaler
func (a *StringArray) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*a = StringArray(arr)
	return nil
}

