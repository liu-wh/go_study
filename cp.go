package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	isR bool
	isF bool
	wg  sync.WaitGroup
)

func cpRoutine(src, dst string) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dstFile.Close()
	bufferRead := bufio.NewReader(srcFile)
	bufferWriter := bufio.NewWriter(dstFile)
	defer bufferWriter.Flush()
	ctx := make([]byte, 1024*1024)
	for {
		n, err := bufferRead.Read(ctx)
		if err != nil {
			break
		}
		bufferWriter.Write(ctx[:n])
	}
	wg.Done()
}

func cp(src, dst string) error {
	srcObj, err := os.Stat(src)
	dstMap := map[string]map[string]string{}
	if err != nil {
		return fmt.Errorf("source file or path not exist")
	}
	if srcObj.IsDir() && !isR {
		return fmt.Errorf("source is a dir, please use -r option")
	}
	if srcObj.IsDir() {
		filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() {
				dstMap[path] = map[string]string{"isDir": "false", "path": path}
			} else {
				if path == src {
					return nil
				}
				dstMap[path] = map[string]string{"isDir": "true", "path": path}
			}
			return nil
		})
	} else {
		dstMap[src] = map[string]string{"isDir": "false", "path": src}
	}
	if len(dstMap) > 1 {
		_, err = os.Stat(dst)
		if err != nil {
			err := os.Mkdir(dst, 755)
			if err != nil {
				return fmt.Errorf("create destnation path  failed")
			}
			for _, v := range dstMap {
				v["path"] = strings.Replace(v["path"], src, "", 1)
			}
		}
		dstAbsPath, _ := filepath.Abs(dst)
		for k, v := range dstMap {
			if v["isDir"] == "true" {
				os.MkdirAll(filepath.Join(dstAbsPath, v["path"]), 0755)
			} else {
				dstFileAbsPath := filepath.Join(dstAbsPath, v["path"])
				_, err = os.Stat(dstFileAbsPath)
				if err == nil && !isF {
					return fmt.Errorf("destnation file already exists, please use -f option")
				}
				dstFileAbsDir := filepath.Dir(dstFileAbsPath)
				_, err = os.Stat(dstFileAbsDir)
				if err != nil {
					os.MkdirAll(dstFileAbsDir, 0755)
				}
				wg.Add(1)
				go cpRoutine(k, dstFileAbsPath)
			}
		}
	} else {
		_, err = os.Stat(dst)
		if err != nil {
			wg.Add(1)
			go cpRoutine(src, dst)
		} else if isF && err == nil {
			wg.Add(1)
			go cpRoutine(src, dst)
		} else {
			return fmt.Errorf("Destination file or dir already exists, please use -f option")
		}
	}
	return nil
}

func main() {
	flag.BoolVar(&isR, "r", false, "Copy a directory recursively")
	flag.BoolVar(&isF, "f", false, "Force copy")
	flag.Parse()
	pathSlice := flag.Args()
	if len(pathSlice) < 2 {
		return
	}
	err := cp(pathSlice[0], pathSlice[1])
	if err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}
