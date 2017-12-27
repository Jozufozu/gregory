package util

import "strings"

var booleans = map[string]bool{
	"yes":   true,
	"true":  true,
	"good":  true,
	"ok":    true,
	"on":    true,
	"no":    false,
	"false": false,
	"bad":   false,
	"off":   false,
}

func GetBool(s string) bool {
	if v, ok := booleans[strings.ToLower(s)]; ok {
		return v
	}
	return false
}
