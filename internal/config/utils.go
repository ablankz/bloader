package config

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func SetNestedValue(v *viper.Viper, key string, value any) {
	keys := strings.Split(key, ".")
	data := v.AllSettings()
	nestedMap := data
	newMap := parse(nestedMap, keys, value)
	v.MergeConfigMap(newMap)
}

func parse(currentMap map[string]any, keys []string, value any) map[string]any {
	k := keys[0]
	remainKeys := keys[1:]
	if strings.Contains(k, "[") {
		idx := strings.Index(k, "[")
		preKey := k[:idx]
		idx2 := strings.Index(k, "]")
		postKey := k[idx+1 : idx2]
		if postKey == "" {
			return currentMap
		}
		index, err := strconv.Atoi(postKey)
		if err != nil {
			return currentMap
		}
		if slice, ok := currentMap[preKey]; ok {
			if _, ok := slice.([]any); ok {
				if len(slice.([]any)) > index {
					if v, ok := slice.([]any)[index].(map[string]any); ok {
						slice.([]any)[index] = parse(v, remainKeys, value)
						currentMap[preKey] = slice
						return currentMap
					} else {
						currentMap[preKey].([]any)[index] = value
						return currentMap
					}
				}
			}
		}
	} else {
		if v, ok := currentMap[k]; ok {
			if mapV, ok := v.(map[string]any); ok {
				currentMap[k] = parse(mapV, remainKeys, value)
				return currentMap
			} else if _, ok := v.([]any); ok {
				return currentMap
			} else {
				currentMap[k] = value
				return currentMap
			}
		} else {
			return currentMap
		}
	}

	return currentMap
}
