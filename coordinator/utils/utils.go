package utils

import (
	"encoding/json"
	"fmt"
)

func TryMarshal(obj any) string {
	if obj == nil {
		return "<nil>"
	}

	bytes, err := json.Marshal(obj)
	if err == nil {
		return string(bytes)
	}

	return fmt.Sprintf("%#v", obj)
}
