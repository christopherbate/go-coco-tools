package data

import (
	"database/sql"
	"fmt"
	"log"
)

type Category struct {
	Name string `json:"name"`
	ID   DBID   `json:"id"`
}

func ListCategories(db *sql.DB) (*[]Category, error) {
	var res = make([]Category, 0, 10)
	rows, err := db.Query("select id, name from categories")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var c Category
		err = rows.Scan(&c.ID, &c.Name)
		res = append(res, c)
	}
	return &res, nil
}

func CreateCategory(db *sql.DB, c *Category) (DBID, error) {
	res, err := db.Exec("insert into categories (name) values (?)", c.Name)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Println("Couldn't retrieve  the last ID.")
		return 0, nil
	}
	return DBID(lastID), nil
}

func GetCategory(db *sql.DB, id DBID) (*Category, error) {
	var ds Category
	err := db.QueryRow("select id, name from categories where id=?", id).Scan(&ds.ID, &ds.Name)

	if err != nil {
		return nil, err
	}

	return &ds, nil
}

func DeleteCategory(db *sql.DB, catId DBID) (int, error) {
	res, err := db.Exec("delete from categories where id=?", catId)
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

func UpdateCategory(db *sql.DB, cat *Category) (int, error) {
	res, err := db.Exec("update categories set name=? where id=?",
		cat.Name, cat.ID)

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
