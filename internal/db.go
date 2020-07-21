package internal

import (
	"database/sql"
	"fmt"
)

const usersTable = "users"

type DB struct {
	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{db: db}
}

func (d *DB) UserExistsInDB(ui *UserInfo) (bool, error) {
	const usersQuery = "SELECT id FROM %s WHERE uuid = $1 and sse_token = $2"

	stmt, err := d.db.Prepare(fmt.Sprintf(usersQuery, usersTable))
	if err != nil {
		return false, err
	}

	defer func() { _ = stmt.Close() }()

	rows, err := stmt.Query(ui.UserID, ui.Nonce)
	if err != nil {
		return false, err
	}

	count, err := d.countRows(rows)

	return count > 0, err
}

func (*DB) countRows(rows *sql.Rows) (int, error) {
	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}
