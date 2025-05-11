package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

func MergeChunksConcurrent(ctx context.Context, chunkDir, outputFile string) error {
	hashes, err := readHashFile(filepath.Join(chunkDir, "hashes.sha256"))
	if err != nil {
		return err
	}

	files, err := os.ReadDir(chunkDir)
	if err != nil {
		return err
	}

	var chunkFiles []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "part_") {
			chunkFiles = append(chunkFiles, file.Name())
		}
	}
	// sort.Strings(chunkFiles)

	sort.Slice(
		chunkFiles, func(i, j int) bool {
			getNum := func(s string) int {
				var num int
				fmt.Sscanf(filepath.Base(s), "part_%d", &num)
				return num
			}
			return getNum(chunkFiles[i]) < getNum(chunkFiles[j])
		},
	)

	type Chunk struct {
		Order int
		Data  []byte
		Err   error
	}
	numWorkers := runtime.NumCPU()
	resultChan := make(chan Chunk, len(chunkFiles))
	sem := semaphore.NewWeighted(int64(numWorkers))
	var wg sync.WaitGroup

	for idx, filename := range chunkFiles {
		wg.Add(1)
		go func(order int, fname string) {
			defer wg.Done()
			sem.Acquire(ctx, 1)
			defer sem.Release(1)

			data, err := os.ReadFile(filepath.Join(chunkDir, fname))
			if err != nil {
				resultChan <- Chunk{order, nil, err}
				return
			}

			expected := hashes[fname]
			actual := sha256.Sum256(data)
			if expected != hex.EncodeToString(actual[:]) {
				resultChan <- Chunk{order, nil, fmt.Errorf("哈希不匹配：%s", fname)}
				return
			}

			resultChan <- Chunk{order, data, nil}
		}(idx, filename)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集并排序
	chunkData := make([][]byte, len(chunkFiles))
	for chunk := range resultChan {
		if chunk.Err != nil {
			return chunk.Err
		}
		chunkData[chunk.Order] = chunk.Data
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, data := range chunkData {
		_, err := outFile.Write(data)
		if err != nil {
			return err
		}
	}

	fmt.Printf("文件成功合并并校验：%s\n", outputFile)

	// ✅ 合并成功后删除分片和哈希文件
	for _, fname := range chunkFiles {
		err := os.Remove(filepath.Join(chunkDir, fname))
		if err != nil {
			fmt.Printf("无法删除分片 %s：%v\n", fname, err)
		}
	}
	err = os.Remove(filepath.Join(chunkDir, "hashes.sha256"))
	if err != nil {
		fmt.Printf("无法删除哈希文件：%v\n", err)
	}

	fmt.Println("原始分片和哈希文件已删除。")
	return nil
}
