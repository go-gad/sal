package sal

import (
	"fmt"
	"regexp"
)

var reQueryArgs = regexp.MustCompile("@[A-Za-z0-9_]+")

// QueryArgs receives the query with named arguments
// and returns a query with posgtresql placeholders and a ordered slice named args.
//
// Naive implementation.
func QueryArgs(query string) (string, []string) {
	var args = make([]string, 0)
	pgQuery := reQueryArgs.ReplaceAllStringFunc(query, func(arg string) string {
		args = append(args, arg[1:])
		return fmt.Sprintf("$%d", len(args))
	})
	return pgQuery, args
}

type KeysIntf map[string]interface{}

type ArgsMap map[string]interface{}

func ProcessQueryAndArgs(query string, reqMap ArgsMap) (string, []interface{}) {
	pgQuery, argsNames := QueryArgs(query)
	var args = make([]interface{}, 0, len(argsNames))
	for _, name := range argsNames {
		args = append(args, reqMap[name])
	}
	return pgQuery, args
}
