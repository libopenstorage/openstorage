package flexvolume

import "encoding/json"

func bytesToJSONOptions(value []byte) (map[string]interface{}, error) {
	if value == nil || len(value) == 0 {
		return nil, nil
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(value, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func jsonOptionsToBytes(jsonOptions map[string]interface{}) ([]byte, error) {
	if jsonOptions == nil || len(jsonOptions) == 0 {
		return nil, nil
	}
	return json.Marshal(jsonOptions)
}
