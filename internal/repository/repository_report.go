package repository

import (
	"database/sql"
	"scim-integrations/internal/repository/query"
)

type ReportRepository interface {
	Select(*query.SelectFilter) ([]*ReportsRow, error)
	Insert(ReportsRow) (int64, error)
}

type reportRepositoryImpl struct{}

func NewReportRepository() ReportRepository {
	return &reportRepositoryImpl{}
}

func (repo *reportRepositoryImpl) Select(filter *query.SelectFilter) ([]*ReportsRow, error) {
	rows, err := execQuery(query.NewReportQuery().Select(filter))
	if err != nil {
		return nil, err
	}
	return repo.reportsFromSQLRows(rows)
}

func (r *reportRepositoryImpl) Insert(report ReportsRow) (int64, error) {
	result, err := exec(
		query.NewReportQuery().Insert(),
		report.StartedAt,
		report.CompletedAt,
		report.CreatedUsersCount,
		report.UpdatedUsersCount,
		report.DeletedUsersCount,
		report.CreatedGroupsCount,
		report.UpdatedGroupsCount,
		report.DeletedGroupsCount,
		report.Succeed,
	)
	if err != nil {
		return 0, err
	}
	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertedID, nil
}

func (*reportRepositoryImpl) reportsFromSQLRows(rows *sql.Rows) ([]*ReportsRow, error) {
	var reports []*ReportsRow
	for rows.Next() {
		report := &ReportsRow{}
		err := rows.Scan(
			&report.ID,
			&report.StartedAt,
			&report.CompletedAt,
			&report.CreatedUsersCount,
			&report.UpdatedUsersCount,
			&report.DeletedUsersCount,
			&report.CreatedGroupsCount,
			&report.UpdatedGroupsCount,
			&report.DeletedGroupsCount,
			&report.Succeed,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return reports, nil
}
