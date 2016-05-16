package common

import "encoding/json"

// ToBytes converts to value to a byte slice.
func ToBytes(val interface{}) ([]byte, error) {
	switch val.(type) {
	case string:
		return []byte(val.(string)), nil
	case []byte:
		b := make([]byte, len(val.([]byte)))
		copy(b, val.([]byte))
		return b, nil
	default:
		return json.Marshal(val)
	}
}
