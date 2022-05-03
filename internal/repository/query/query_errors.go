package query

type ErrorsQuery interface {
	Select(filter *SelectFilter) string
	Insert() string
}
type errorsQueryImpl struct{}

func NewErrorsQuery() ErrorsQuery {
	return &errorsQueryImpl{}
}

func (*errorsQueryImpl) Select(filter *SelectFilter) string {
	return addQueryFilters("SELECT * FROM errors", filter)
}

func (*errorsQueryImpl) Insert() string {
	return `
		INSERT INTO errors (
			occurred_time,
			error
		) VALUES ($1, $2);
	`
}
