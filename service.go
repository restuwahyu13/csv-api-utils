package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
)

type InterfaceService interface {
	Merge(req *CSVMergePayload) ([][]string, error)
	Split(req *CSVSplitPayload) ([][]string, error)
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

func (s *StructService) Merge(req *CSVMergePayload) ([][]string, error) {
	var (
		poolSize int           = 1000
		headers  []string      = []string{}
		mutex    *sync.RWMutex = &sync.RWMutex{}
		dir      fs.FS         = os.DirFS(req.InputDir)
	)

	/*
	* ================================
	* GET CSV FILES HEADERS
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

	for _, v := range files[0:1] {
		r, err := fs.ReadFile(dir, v)

		if err != nil {
			defer log.Println(err.Error())
			return nil, err
		}

		reader := csv.NewReader(bytes.NewReader(r))
		metadata, err := reader.Read()

		if err != nil {
			defer log.Println(err.Error())
			return nil, err
		}

		mutex.Lock()
		headers = append(headers, metadata...)
		mutex.Unlock()
	}

	/*
	* ================================
	* GET CSV FILES WITHOUT HEADERS
	* ================================
	 */

	contents := [][]string{headers}

	for _, v := range files {
		r, err := fs.ReadFile(dir, v)

		if err != nil {
			defer log.Println(err.Error())
			return nil, err
		}

		reader := csv.NewReader(bytes.NewReader(r))
		records, err := reader.ReadAll()

		if err != nil {
			return nil, err
		}

		for _, v := range records[1:] {
			mutex.Lock()
			contents = append(contents, v)
			mutex.Unlock()
		}
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
	},
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
	)

	if err := pool.Invoke(contents); err != nil {
		defer log.Println(err.Error())
		return nil, err
	}

	defer pool.Release()
	return contents, nil
}

/*
* ================================
* SPLIT SERVICE
* ================================
 */

func (s *StructService) Split(req *CSVSplitPayload) ([][]string, error) {
	var (
		poolSize     int           = 1000
		csvChunks    [][]int       = [][]int{}
		csvChunkSize int           = req.PerRow
		csvIndex     []int         = []int{}
		firstIndex   []int         = []int{}
		lastIndex    []int         = []int{}
		mutex        *sync.RWMutex = &sync.RWMutex{}
		safeMap      *atomic.Value = &atomic.Value{}
	)

	/*
	* ================================
	* READ CSV FILES HEADERS
	* ================================
	 */

	if _, err := os.Stat(req.InputFile); err != nil {
		defer log.Println(err.Error())
		return nil, errors.New("File not found")
	}

	readFile, err := os.ReadFile(req.InputFile)
	if err != nil {
		return nil, err
	}

	csvRead := csv.NewReader(bytes.NewBuffer(readFile))
	reader, err := csvRead.ReadAll()
	if err != nil {
		return nil, err
	}

	csvHeader := reader[0]
	csvContent := reader[1:]
	csvLength := len(reader)

	/*
	* ================================
	* GET CSV FILES CONTENT
	* ================================
	 */

	for j := 1; j <= csvLength; j++ {
		csvIndex = append(csvIndex, j)
	}

	for i := 0; i < len(csvIndex); i += csvChunkSize {
		end := i + csvChunkSize

		if end > len(csvIndex) {
			end = len(csvIndex)
		}

		mutex.Lock()
		csvChunks = append(csvChunks, csvIndex[i:end])
		mutex.Unlock()
	}

	for _, v := range csvChunks {
		mutex.Lock()
		firstIndex = append(firstIndex, v[0]-1)
		mutex.Unlock()
	}

	for _, v := range csvChunks {
		mutex.Lock()
		lastIndex = append(lastIndex, v[len(v)-1])
		mutex.Unlock()
	}

	/*
	* ================================
	* WRITE CSV FILES
	* ================================
	 */

	pool, err := ants.NewPoolWithFunc(poolSize, func(content interface{}) {
		data := content
		safeMap.Store(data)
		safeData := safeMap.Load().(map[string]interface{})

		fileName := fmt.Sprintf("%d-%s-%s.csv", safeData["index"], req.Filename, time.Now().Format("2006-01-02"))
		outputDir := req.OutputFile + "/" + fileName

		outputFile, err := os.Create(outputDir)

		if err != nil {
			log.Fatal(err)
			return
		}

		write := csv.NewWriter(outputFile)

		if err := write.WriteAll(safeData["content"].([][]string)); err != nil {
			log.Fatal(err.Error())
			return
		}

		defer func() {
			outputFile.Close()
			write.Flush()
		}()
	},
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
	)

	for i, v := range firstIndex {
		firstIndex := v
		lastIndex := lastIndex[i]
		csvContents := [][]string{csvHeader}

		mutex.Lock()
		csvContents = append(csvContents, csvContent[firstIndex:lastIndex]...)
		mutex.Unlock()

		data := make(map[string]interface{})
		data["index"] = i + 1
		data["content"] = csvContents

		if err := pool.Invoke(data); err != nil {
			defer log.Println(err.Error())
			return nil, err
		}
	}

	defer pool.Release()
	return nil, nil
}
