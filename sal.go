package sal

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var reQueryArgs = regexp.MustCompile("@[a-z]+")

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

// CamelCaseToUnderscore converts from camel case form to underscore separated form.
// Ex.: MyFunc => my_func
func CamelCaseToUnderscore(str string) string {
	var output []rune
	var segment []rune
	for _, r := range str {

		// not treat number as separate segment
		if !unicode.IsLower(r) && string(r) != "_" && !unicode.IsNumber(r) {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}

// UnderscoreToCamelCase converts from underscore separated form to camel case form.
// Ex.: my_func => MyFunc
func UnderscoreToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}
