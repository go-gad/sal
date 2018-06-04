package sal

// QueryArgs receives the query with named arguments
// and returns a query with posgtresql placeholders and a ordered slice named args.
func QueryArgs(query string) (string, []string) {
	return "", []string{}
}
