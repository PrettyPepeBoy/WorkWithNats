package validation

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

func Valid(data []byte) bool {
	var dt interface{}
	err := json.Unmarshal(data, &dt)
	if err != nil {
		logrus.Errorf("failed to unmarshal json, error: %v", err)
		return false
	}
	res, ok := dt.(map[string]interface{})
	if !ok {
		return false
	}
	if len(res) == 0 {
		return false
	}
	if len(res) > 10 {
		return false
	}
	for _, val := range res {
		if !castType(val) {
			return false
		}
	}
	return true
}

func castType(val interface{}) bool {
	switch v := val.(type) {
	case float64:
		valInt := int(v)
		if valInt <= 0 {
			return false
		}
		if valInt > (2 << 32) {
			return false
		}
		return true
	case string:
		if len(v) > 20 {
			return false
		}
		if len(v) == 0 {
			return false
		}
		return correctString(v)
	case bool:
		return true
	default:
		return false
	}
}

func correctString(v string) bool {
	for _, elem := range v {
		if !((elem >= 'a' && elem <= 'z') || (elem >= 'A' && elem <= 'Z')) {
			return false
		}
	}
	return true
}
