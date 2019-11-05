package main

import (
	"coco_tools/DLFS"
	"coco_tools/data"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func buildImagesTable(fbsFilename string, sqliteFilename string) error {
	// Open a connection to the sqlite database
	db, err := data.OpenDB(fmt.Sprintf("file:%s?cache=private&_sync=0", sqliteFilename))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer db.Close()

	// Open the flatbuffer file.
	buf, err := ioutil.ReadFile(fbsFilename)
	fbsDataset := DLFS.GetRootAsDataset(buf, 0)
	counts := make([]int, 2)

	stmt, err := db.Prepare("insert into images (filename) values(?)")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < fbsDataset.ExamplesLength(); i++ {
		ex := new(DLFS.Example)
		if fbsDataset.Examples(ex, i) {
			_, err := stmt.Exec(string(ex.FileName()))
			if err != nil {
				counts[1]++
				continue
			}
			counts[0]++
		}
	}

	fmt.Println(counts)
	return nil
}

func main() {
	if len(os.Args) < 4 {
		usage := fmt.Sprintf("Usage: %s localDB fbsFile dataDir", os.Args[0])
		fmt.Println(usage)
		return
	}

	dbFile := os.Args[1]
	fbsFile := os.Args[2]

	// buildImagesTable(fbsFile, dbFile)
	fmt.Println(dbFile, fbsFile)

	return

}
