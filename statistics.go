package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	path      string
	ext       string
	totalLine int
	lineMap   map[string]int
	channel   chan int
	wg        sync.WaitGroup
	wg1       sync.WaitGroup
	trimSpace bool
)

func statistics(path string) {
	line := 0
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	bufferFile := bufio.NewReader(file)
	for {
		ctx, _, err := bufferFile.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}
		lineString := string(ctx)
		lineString = strings.TrimSpace(lineString)
		if trimSpace && lineString == "" {
			continue
		}
		line++
	}
	channel <- line
	wg.Done()
}

func iterPath(path string, info fs.FileInfo, err error) error {
	_ext := filepath.Ext(path)
	if info.IsDir() || _ext != ext {
		return nil
	}
	wg.Add(1)
	go statistics(path)
	// lineMap[path] = line

	return nil
}

func statisticsChannel() {
	for line := range channel {
		totalLine += line
	}
	wg1.Done()
}

func main() {
	lineMap = map[string]int{}
	channel = make(chan int, 10)
	flag.StringVar(&path, "p", ".", "path")
	flag.StringVar(&ext, "t", ".go", "file type")
	flag.BoolVar(&trimSpace, "s", false, "whether blank lines are counted")
	flag.Parse()
	wg1.Add(1)
	go statisticsChannel()
	err := filepath.Walk(path, iterPath)
	if err != nil {
		fmt.Println(err)
	}
	wg.Wait()
	close(channel)
	wg1.Wait()
	fmt.Println(totalLine)
	// fmt.Println(lineMap)
}
