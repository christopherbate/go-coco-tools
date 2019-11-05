package data

import (
	"database/sql"
	"fmt"
)

type Dataset struct {
	ID   DBID   `json:"id"`
	Name string `json:"name"`
}

func ListDatasets(db *sql.DB) (*[]Dataset, error) {
	var res = make([]Dataset, 0, 10)
	rows, err := db.Query("select id, name from datasets")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var c Dataset
		err = rows.Scan(&c.ID, &c.Name)
		res = append(res, c)
	}
	return &res, nil
}

func GetDataset(db *sql.DB, id DBID) (*Dataset, error) {
	var ds Dataset
	err := db.QueryRow("select id, name from datasets where id=?", id).Scan(&ds.ID, &ds.Name)

	if err != nil {
		return nil, err
	}

	return &ds, nil
}

func CreateDataset(db *sql.DB, c *Dataset) (DBID, error) {
	res, err := db.Exec("insert into datasets (name) values (?)", c.Name)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return DBID(lastID), nil
}

func DeleteDataset(db *sql.DB, dsID DBID) (int, error) {
	res, err := db.Exec("delete from datasets where id=?", dsID)
	if err != nil {
		return 0, nil
	}

	numModified, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return 1, nil
	}

	if numModified > 0 {
		return 1, nil
	}
	return 0, nil
}

func UpdateDataset(db *sql.DB, ds *Dataset) (int, error) {
	res, err := db.Exec("update datasets set name=? where id=?",
		ds.Name, ds.ID)

	if err != nil {
		return 0, err
	}

	numAffected, err := res.RowsAffected()

	if err != nil {
		return 0, err
	}

	if numAffected > 0 {
		return 1, nil
	}

	return 0, nil
}
