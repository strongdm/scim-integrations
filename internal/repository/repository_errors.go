package repository

import (
	"database/sql"
	"scim-integrations/internal/repository/query"
)

type ErrorsRepository interface {
	Select(*query.SelectFilter) ([]ErrorsRow, error)
	Insert(ErrorsRow) (int64, error)
}
type errorsRepositoryImpl struct{}

func NewErrorsRepository() ErrorsRepository {
	return &errorsRepositoryImpl{}
}

func (repo *errorsRepositoryImpl) Select(filter *query.SelectFilter) ([]ErrorsRow, error) {
	rows, err := execQuery(query.NewErrorsQuery().Select(filter))
	if err != nil {
		return nil, err
	}
	return repo.errorsFromSQLRows(rows)
}

func (*errorsRepositoryImpl) Insert(row ErrorsRow) (int64, error) {
	result, err := exec(
		query.NewErrorsQuery().Insert(),
		row.Err,
		row.OccurredTime,
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

func (*errorsRepositoryImpl) errorsFromSQLRows(rows *sql.Rows) ([]ErrorsRow, error) {
	var errRows []ErrorsRow
	for rows.Next() {
		row := ErrorsRow{}
		err := rows.Scan(
			&row.ID,
			&row.Err,
			&row.OccurredTime,
		)
		if err != nil {
			return nil, err
		}
		errRows = append(errRows, row)
	}
	return errRows, nil
}
