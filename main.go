package main

import (
	"coco_tools/DLFS"
	"coco_tools/data"
	"database/sql"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type dataset struct {
	db         *sql.DB
	fbData     *DLFS.Dataset
	dataDir    string
	outputPath string
}

func (ds *dataset) Open(fbsFilename string, sqliteFilename string, dataDir string) error {
	// Open a connection to the sqlite database
	db, err := data.OpenDB(fmt.Sprintf("file:%s?cache=shared&_sync=0", sqliteFilename))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	ds.db = db
	ds.dataDir = dataDir

	// Open the flatbuffer file.
	buf, err := ioutil.ReadFile(fbsFilename)
	ds.fbData = DLFS.GetRootAsDataset(buf, 0)

	return nil
}

func (ds *dataset) buildCategoriesTable() error {
	numCats := ds.fbData.CategoriesLength()
	for idx := 0; idx < numCats; idx++ {
		cat := new(DLFS.Category)
		if ds.fbData.Categories(cat, idx) {
			_, err := ds.db.Exec("insert into categories (id, name, cocoid) values (?,?,?)", idx, cat.Name(), cat.Id())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ds *dataset) buildImagesTable() error {
	counts := make([]int, 2)
	stmt, err := ds.db.Prepare("insert into images (filename) values(?)")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < ds.fbData.ExamplesLength(); i++ {
		ex := new(DLFS.Example)
		if ds.fbData.Examples(ex, i) {
			_, err := stmt.Exec(string(ex.FileName()))
			if err != nil {
				counts[1]++
				continue
			}
			counts[0]++
		}
	}

	fmt.Println("Insertion results: ", counts)
	return nil
}

func (ds *dataset) createThumbnails(exampleIdx int, outputPath string) error {
	ex := new(DLFS.Example)
	if ds.fbData.Examples(ex, exampleIdx) {
		filename := path.Join(ds.dataDir, string(ex.FileName()))
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		imgData, err := jpeg.Decode(file)
		if err != nil {
			log.Println(filename)
			return err
		}
		file.Close()

		// Get the annotations
		numAnn := ex.AnnotationsLength()
		for i := 0; i < numAnn; i++ {
			// Load the ann
			ann := new(DLFS.Annotation)
			ex.Annotations(ann, i)

			if ann == nil {
				return fmt.Errorf("Could not retrieve annotation")
			}

			// Convert source rectangle to dst image space
			var bbox *DLFS.BoundingBox = ann.Bbox(nil)
			imgBox := image.Rectangle{image.Point{int(bbox.X1()), int(bbox.Y1())},
				image.Point{int(bbox.X2()), int(bbox.Y2())}}

			// Create image
			tbImage := image.NewRGBA(imgBox)
			draw.Draw(tbImage, tbImage.Bounds(), imgData, imgBox.Bounds().Min, draw.Src)

			tbFilename := path.Join(outputPath, fmt.Sprintf("%d-%d-tb.jpg", ex.Id(), i))
			outfile, err := os.Create(tbFilename)
			if err != nil {
				return err
			}

			err = jpeg.Encode(outfile, tbImage, nil)
			if err != nil {
				return err
			}
			outfile.Close()

			_, err = ds.db.Exec("insert into thumbnails (filename, category, example_idx, width, height, area) values (?,?,?,?,?,?)", tbFilename, ann.CatId(),
				ex.Idx(), imgBox.Size().X, imgBox.Size().Y, imgBox.Size().X*imgBox.Size().Y)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("No such example")
}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}

func main() {
	if len(os.Args) < 4 {
		usage := fmt.Sprintf("Usage: %s localDB fbsFile dataDir", os.Args[0])
		fmt.Println(usage)
		return
	}

	dbFile := os.Args[1]
	fbsFile := os.Args[2]
	dataDir := os.Args[3]

	startTime := time.Now()

	var ds dataset
	err := ds.Open(fbsFile, dbFile, dataDir)
	if err != nil {
		log.Fatal(err)
	}

	parallelism := 10
	idxChannel := make(chan int, parallelism)
	errorChannel := make(chan error, parallelism)
	var wg sync.WaitGroup
	numImgs := ds.fbData.ExamplesLength()

	go func(total int) {
		for j := 0; j < 100000; j++ {
			idxChannel <- j
		}
		close(idxChannel)
	}(numImgs)

	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func(idx int, ds *dataset) {
			defer wg.Done()
			for idx := range idxChannel {
				err := ds.createThumbnails(idx, "/home/chris/datasets/coco/thumbnails")
				errorChannel <- err
			}
		}(i, &ds)
	}

	go func(total int) {
		doneCnt := 0
		errCnt := 0
		spinString := string(`-\|/`)
		spinIdx := 0
		for err := range errorChannel {
			if err != nil {
				errCnt++
				log.Println(err)
			} else {
				doneCnt++
			}
			fmt.Printf("\r%c %d/%d done %d errors", spinString[spinIdx%len(spinString)], doneCnt, total, errCnt)
			spinIdx++
		}

		fmt.Println("\n Limit reached.")
	}(numImgs)

	wg.Wait()

	close(errorChannel)

	endTime := time.Now()
	elapsed := endTime.Sub(startTime)
	log.Printf("Done %.3f elapsd", elapsed.Seconds())

	time.Sleep(time.Second * 1)
}
