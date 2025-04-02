package functions

import (
	"strconv"
	"sync"
)

var DefaultCenterOptions = map[string]string{
	"service_jwt_valid_time": "30",
}

var Options sync.Map

func GetOption(keyName string) string {
	if v, ok := Options.Load(keyName); ok {
		return v.(string)
	} else {
		return ""
	}
}

func SetOption[T ~string | ~bool | ~int](keyName string, value T) error {
	newValue := ""
	switch any(value).(type) {
	case string:
		newValue = any(value).(string)
	case bool:
		if any(value).(bool) {
			newValue = "1"
		} else {
			newValue = "0"
		}
	case int:
		newValue = strconv.Itoa(any(value).(int))
	}

	v, ok := Options.Load(keyName)
	if ok && v == newValue {
		return nil
	}

	Options.Store(keyName, newValue)

	return nil
}

func DeleteOption(keyName string) error {
	Options.Delete(keyName)
	return nil
}

func InitOptions() {
	optionsKV := make(map[string]string)

	for key, value := range DefaultCenterOptions {
		if v, ok := optionsKV[key]; ok {
			Options.Store(key, v)
		} else {
			Options.Store(key, value)
		}
	}
}
