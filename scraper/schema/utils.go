package schema

import (
	"log"
	"strings"
	"time"

	"github.com/sosodev/duration"

	"github.com/astappiev/microdata"
	"github.com/borschtapp/krip/utils"
)

func getStringOrItem(val interface{}) (string, *microdata.Item) {
	switch v := val.(type) {
	case string:
		return utils.CleanupInline(v), nil
	case *microdata.Item:
		return "", v
	default:
		log.Printf("unable to process `%s`, unexpected type `%T`\n", val, val)
	}

	return "", nil
}

func getStringOrChild(val interface{}, child ...string) (string, bool) {
	if text, item := getStringOrItem(val); len(text) != 0 {
		return text, true
	} else if item != nil {
		return getPropertyString(item, child...)
	}
	return "", false
}

func getPropertyStringOrChild(item *microdata.Item, key string, child ...string) (string, bool) {
	if val, ok := item.GetProperty(key); ok {
		return getStringOrChild(val, child...)
	}

	return "", false
}

func getPropertyString(item *microdata.Item, key ...string) (string, bool) {
	if val, ok := item.GetProperty(key...); ok {
		if text, ok := val.(string); ok {
			return text, len(text) != 0
		} else {
			log.Printf("unable to retrieve `string` value of `%s` in (%v)\n", key, item)
		}
	}

	return "", false
}

func getPropertyInt(item *microdata.Item, key ...string) (int, bool) {
	if val, ok := item.GetProperty(key...); ok {
		switch v := val.(type) {
		case *microdata.Item:
			return getPropertyInt(v, "value")
		default:
			return utils.FindInt(val), true
		}
	}

	return 0, false
}

func getPropertyFloat(item *microdata.Item, key ...string) (float64, bool) {
	if val, ok := item.GetProperty(key...); ok {
		return utils.FindFloat(val), true
	}

	return 0, false
}

func getPropertiesArray(item *microdata.Item, keys ...string) ([]string, bool) {
	if values, ok := item.GetProperties(keys...); ok {
		var arr []string
		for _, val := range values {
			if val, ok := val.(string); ok && len(val) != 0 {
				arr = append(arr, val)
			}
		}

		return arr, len(arr) != 0
	}

	return nil, false
}

func getPropertiesKeywords(item *microdata.Item, keys ...string) ([]string, bool) {
	if values, ok := getPropertiesArray(item, keys...); ok {
		var arr []string

		if len(values) == 1 && strings.Contains(values[0], ",") {
			values = strings.Split(values[0], ",")
		}

		for _, text := range values {
			if text := utils.CleanupInline(text); len(text) != 0 {
				arr = append(arr, text)
			}
		}

		return utils.Deduplicate(arr), len(arr) != 0
	}

	return nil, false
}

func getPropertyDuration(item *microdata.Item, key ...string) (time.Duration, bool) {
	if val, ok := getPropertyString(item, key...); ok && len(val) != 0 {
		if d, err := duration.Parse(utils.RemoveSpaces(val)); err == nil {
			return d.ToTimeDuration(), true
		} else if val, ok := utils.ParseDuration(val); ok {
			return val, true
		} else {
			log.Printf("unable to parse duration `%s`: %s\n", val, err.Error())
		}
	}

	return 0, false
}
