package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/panjf2000/ants/v2"
)

type InterfaceService interface {
	Merge(req *CSVPayload) ([][]string, error)
}

type StructService struct{}

func NewService() InterfaceService {
	return &StructService{}
}

/*
* ================================
* MERGE SERVICE
* ================================
 */

func (s *StructService) Merge(req *CSVPayload) ([][]string, error) {
	var (
		poolSize int           = 5000
		headers  []string      = []string{}
		contents []StructInput = []StructInput{}
		content  StructInput   = StructInput{}
		mutex    *sync.RWMutex = &sync.RWMutex{}
		dir      fs.FS         = os.DirFS(req.InputDir)
	)

	/*
	* ================================
	* GET CSV FILES
	* ================================
	 */

	files, err := fs.Glob(dir, "*.csv")
	if err != nil {
		defer log.Println(err)
		return nil, err
	}

	filesLength := len(files)

	if filesLength <= 0 {
		return nil, errors.New("CSV file not found")
	}

	csvFiles := files[0:1]

	/*
	* ================================
	* GET CSV FILES HEADERS
	* ================================
	 */

	for _, v := range csvFiles {
		r, err := fs.ReadFile(dir, v)
		if err != nil {
			defer log.Println(err)
			return nil, err
		}

		reader := csv.NewReader(bytes.NewReader(r))
		metadata, err := reader.Read()

		if err != nil {
			defer log.Println(err)
			return nil, err
		}

		mutex.Lock()
		headers = append(headers, metadata...)
		mutex.Unlock()

		break
	}

	/*
	* ================================
	* GET CSV FILES WITH HEADERS
	* ================================
	 */

	for _, v := range files[0:1] {
		r, err := fs.ReadFile(dir, v)
		if err != nil {
			defer log.Println(err)
			return nil, err
		}

		reader := csv.NewReader(bytes.NewReader(r))
		decoder, err := csvutil.NewDecoder(reader)

		if err != nil {
			defer log.Println(err)
			return nil, err
		}

		for {
			if err := decoder.Decode(&content); err == io.EOF {
				break
			} else if err != nil {
				defer log.Println(err)
				return nil, err
			}

			mutex.Lock()
			contents = append(contents, content)
			mutex.Unlock()
		}
	}

	/*
	* ================================
	* GET CSV FILES WITHOUT HEADERS
	* ================================
	 */

	for _, v := range files[1:filesLength] {
		r, err := fs.ReadFile(dir, v)
		if err != nil {
			defer log.Println(err)
			return nil, err
		}

		reader := csv.NewReader(bytes.NewReader(r))
		decoder, err := csvutil.NewDecoder(reader)

		if err != nil {
			defer log.Println(err)
			return nil, err
		}

		for {
			if err := decoder.Decode(&content); err == io.EOF {
				break
			} else if err != nil {
				defer log.Println(err)
				return nil, err
			}

			mutex.Lock()
			contents = append(contents, content)
			mutex.Unlock()
		}
	}

	contentsByte, err := csvutil.Marshal(&contents)
	if err != nil {
		defer log.Println(err)
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(contentsByte))
	csvData, err := reader.ReadAll()

	if err != nil {
		defer log.Println(err)
		return nil, err
	}

	/*
	* ================================
	* WRITE CSV FILES
	* ================================
	 */

	pool, err := ants.NewPoolWithFunc(poolSize, func(data interface{}) {
		fileName := fmt.Sprintf("%s-%s.csv", req.Filename, time.Now().Format("2006-01-02"))
		outputDir := req.OutputDir + "/" + fileName

		outputFile, err := os.Create(outputDir)

		if err != nil {
			log.Fatal(err)
			return
		}

		write := csv.NewWriter(outputFile)

		if err := write.WriteAll(data.([][]string)); err != nil {
			log.Fatal(err.Error())
			return
		}

		defer func() {
			outputFile.Close()
			write.Flush()
		}()
	})

	if err := pool.Invoke(csvData); err != nil {
		return nil, err
	}

	defer pool.Release()
	return csvData, nil
}
