package query

type ReportQuery interface {
	Select(*SelectFilter) string
	Insert() string
}
type reportQueryImpl struct{}

func NewReportQuery() ReportQuery {
	return &reportQueryImpl{}
}

func (*reportQueryImpl) Select(filter *SelectFilter) string {
	return addQueryFilters("SELECT * FROM reports", filter)
}

func (*reportQueryImpl) Insert() string {
	return `
		INSERT INTO reports (
			start_time,
			end_time,
			users_created,
			users_updated,
			users_deleted,
			groups_created,
			groups_updated,
			groups_deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
}
