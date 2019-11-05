package data

import (
	"database/sql"
	"fmt"
	"log"
)

type Model struct {
	ID        DBID   `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	DatasetID DBID   `json:"datasetId"`
}

type ModelPerformance struct {
	ID         DBID       `json:"id"`
	ModelID    DBID       `json:"modelId"`
	Classes    ClassNames `json:"classNames"`
	SavedModel string     `json:"savedModel"`
}

type ClassModelPerf struct {
	ModelPerformance
	Precision float32 `json:"precision"`
	Recall    float32 `json:"recall"`
}

type ClassNames []string

type DetectionModelPerf struct {
	ModelPerformance
	AvgPrecision float32 `json:"avgPrecision"`
	AvgRecall    float32 `json:"avgRecall"`
}

func ListModels(db *sql.DB) (*[]Model, error) {
	var res = make([]Model, 0, 10)
	rows, err := db.Query("select id, name, type, datasetid from models")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var c Model
		err = rows.Scan(&c.ID, &c.Name, &c.Type, &c.DatasetID)
		res = append(res, c)
	}
	return &res, nil
}

func GetModel(db *sql.DB, id DBID) (*Model, error) {
	var model Model
	err := db.QueryRow("select id, name, type, datasetid from models where id=?", id).Scan(&model.ID,
		&model.Name, &model.Type, &model.DatasetID)

	if err != nil {
		return nil, err
	}

	return &model, nil
}

func CreateModel(db *sql.DB, c *Model) (DBID, error) {
	res, err := db.Exec("insert into models (name, type, datasetid) values (?,?,?)",
		c.Name, c.Type, c.DatasetID)
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

func DeleteModel(db *sql.DB, id DBID) (int, error) {
	res, err := db.Exec("delete from models where id=?", id)
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

func UpdateModel(db *sql.DB, req *Model) (int, error) {
	res, err := db.Exec("update models set name=?, type=?, datasetid=? where id=?",
		req.Name, req.Type, req.DatasetID, req.ID)

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
