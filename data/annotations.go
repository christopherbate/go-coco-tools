package data

import (
	"database/sql"
	"fmt"
)

type Annotation struct {
	ID         DBID
	ImageID    DBID
	Box        Box
	CategoryID DBID
}

func ListAnnotationsForImage(db *sql.DB, image *Image) (*[]Annotation, error) {
	var res = make([]Annotation, 0, 1)

	if image != nil {
		rows, err := db.Query("select id, box, categoryid, imageid from annotations where imageid = ?",
			image.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var i Annotation
			err := rows.Scan(&i.ID, &i.Box, &i.CategoryID, &i.ImageID)
			if err != nil {
				return nil, err
			}
			res = append(res, i)
		}

		return &res, nil
	}

	return nil, fmt.Errorf("No image provided.")
}

func ListAllAnnotations(db *sql.DB) (*[]Annotation, error) {
	var res = make([]Annotation, 0, 10)
	rows, err := db.Query("select id, box, categoryid, imageid from annotations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var i Annotation
		err := rows.Scan(&i.ID, &i.Box, &i.CategoryID, &i.ImageID)
		if err != nil {
			return nil, err
		}
		res = append(res, i)
	}

	return &res, nil
}

func CreateAnnotation(db *sql.DB, ann *Annotation) (DBID, error) {
	res, err := db.Exec("insert into annotations (box, categoryid, imageid) values (?,?,?)",
		ann.Box, ann.CategoryID, ann.ImageID)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return DBID(lastID), nil
}

func UpdateAnnotations(db *sql.DB, anns *[]Annotation) (int, error) {
	var numUpdate int = 0
	for _, ann := range *anns {
		_, err := db.Exec("update annotations set categoryid=?, box=? where id=?",
			ann.CategoryID, ann.Box, ann.ID)
		if err != nil {
			return numUpdate, err
		}
		numUpdate++
	}
	return numUpdate, nil
}

func CreateAnnotations(db *sql.DB, anns *[]Annotation) (int, error) {
	var numCreate int = 0
	for _, ann := range *anns {
		_, err := CreateAnnotation(db, &ann)
		if err != nil {
			return numCreate, err
		}
		numCreate++
	}
	return numCreate, nil
}

func DeleteAnnotations(db *sql.DB, deletionIDs *[]DBID) (int, error) {
	var numDelete int = 0
	for _, delID := range *deletionIDs {
		_, err := db.Exec("delete from annotations where id = ?", delID)
		if err != nil {
			return numDelete, err
		}
		numDelete++
	}
	return numDelete, nil
}
