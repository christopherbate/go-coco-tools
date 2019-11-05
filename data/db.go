package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(conn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type DBID int64

type Box [4]int

func (box *Box) Scan(src interface{}) error {
	val, ok := src.(string)
	if !ok {
		return fmt.Errorf("unable to convert src to string")
	}
	bbox := Box{0, 0, 0, 0}
	for idx, i := range strings.Split(val, ",") {
		j, err := strconv.Atoi(i)
		if err != nil {
			return err
		}
		bbox[idx] = j		
	}
	*box = bbox
	return nil
}

func (box Box) Value() (driver.Value, error) {
	var boxString string
	boxString = strings.Trim(strings.Join(strings.Split(fmt.Sprint(box), " "), ","), "[]")
	return driver.Value(boxString), nil
}
