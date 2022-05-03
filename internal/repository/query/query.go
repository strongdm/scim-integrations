package query

import "fmt"

type SelectFilter struct {
	Limit   int
	OrderBy string
}

func addQueryFilters(query string, filter *SelectFilter) string {
	filteredQuery := query
	if filter != nil {
		if filter.OrderBy != "" {
			filteredQuery = fmt.Sprintf("%s order by %s", filteredQuery, filter.OrderBy)
		}
		if filter.Limit > 0 {
			filteredQuery = fmt.Sprintf("%s limit %d", filteredQuery, filter.Limit)
		}
	}
	return fmt.Sprintf("%s;", filteredQuery)
}
