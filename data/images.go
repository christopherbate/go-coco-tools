package data

import (
	"database/sql"
	"fmt"
	"log"
)

type Image struct {
	Filename    string       `json:"filename"`
	ID          DBID         `json:"id"`
	Annotations []Annotation `json:"annotations"`
}

type ImageDB struct {
	imgs map[DBID]Image
}

type ImageFilter struct {
	CatID   DBID   `json:"catId"`
	Labeled string `json:"labeled"`
}

func CreateImageTestDB() *ImageDB {
	var db ImageDB
	db.imgs = make(map[DBID]Image)
	db.imgs[1] = Image{Filename: "images/n02096051_7516.jpg", ID: 1}
	db.imgs[2] = Image{Filename: "images/solar_88.jpg", ID: 2}
	return &db
}

var testImgDB = CreateImageTestDB()

func ListImages(db *sql.DB, labelFilter string,
	categoryFilter DBID) (*[]Image, error) {
	var res = make([]Image, 0, 10)
	rows, err := db.Query("select id, filename from images")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var img Image
		err = rows.Scan(&img.ID, &img.Filename)
		res = append(res, img)
	}
	return &res, nil
}

func GetImage(db *sql.DB, imageID DBID) (*Image, error) {
	var result Image
	err := db.QueryRow("select id, filename from images where id = ? limit 1",
		imageID).Scan(&result.ID, &result.Filename)
	if err != nil {
		return nil, err
	}

	anns, err := ListAnnotationsForImage(db, &result)
	if err != nil {
		return nil, err
	}

	result.Annotations = *anns

	return &result, nil
}

func CreateImage(db *sql.DB, image *Image) (DBID, error) {
	res, err := db.Exec("insert into images (filename) values (?)", image.Filename)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Println("Could not get last image inserted id")
		return 0, nil
	}
	return DBID(lastID), nil
}

func DeleteImage(db *sql.DB, ID DBID) (int, error) {
	res, err := db.Exec("delete from images where id=?", ID)
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

func UpdateImage(db *sql.DB, img *Image) (int, error) {
	res, err := db.Exec("update images set filename=? where id=?",
		img.Filename, img.ID)

	if err != nil {
		return 0, err
	}

	numAffected, err := res.RowsAffected()

	if err != nil {
		return 0, err
	}

	numUpdate, err := UpdateAnnotations(db, &img.Annotations)
	if err != nil {
		return 1, err
	}

	fmt.Println("Updated anns", numUpdate)
	fmt.Println(img.Annotations)

	if numAffected > 0 {
		return 1, nil
	}

	return 0, nil
}
