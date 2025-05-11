package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

type ChunkResult struct {
	Filename string
	Hash     string
	Err      error
}

func SplitFileConcurrent(ctx context.Context, filePath, outputDir string, chunkSize int64) error {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer inputFile.Close()
	// 获取文件信息以计算总大小和分片数量
	info, err := inputFile.Stat()
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}
	fileSize := info.Size()
	partSize := PartSize
	if chunkSize > PartSize {
		partSize = chunkSize
	}
	numParts := int(fileSize / partSize)
	if fileSize%partSize != 0 {
		numParts++
	}
	// basePath := filepath.Dir(outputDir)
	basePath := outputDir
	// 并发限制
	numWorkers := runtime.NumCPU()

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(numWorkers))
	resultChan := make(chan ChunkResult)
	buffer := make([]byte, partSize)
	for i := 0; i < numParts; i++ {
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// 拷贝当前块数据（避免后续被覆盖）
		chunkData := make([]byte, n)
		copy(chunkData, buffer[:n])
		chunkName := fmt.Sprintf("part_%d", i)
		wg.Add(1)

		go func(data []byte, name string) {
			defer wg.Done()
			sem.Acquire(ctx, 1)
			defer sem.Release(1)

			err := os.WriteFile(filepath.Join(basePath, name), data, 0644)
			if err != nil {
				resultChan <- ChunkResult{name, "", err}
				return
			}

			hash := sha256.Sum256(data)
			resultChan <- ChunkResult{name, hex.EncodeToString(hash[:]), nil}
		}(chunkData, chunkName)
	}

	// 收集所有结果
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 写入 hash 文件
	hashFile, err := os.Create(filepath.Join(basePath, "hashes.sha256"))
	if err != nil {
		return err
	}
	defer hashFile.Close()

	for result := range resultChan {
		if result.Err != nil {
			return fmt.Errorf("分片 %s 出错: %v", result.Filename, result.Err)
		}
		_, err := fmt.Fprintf(hashFile, "%s  %s\n", result.Hash, result.Filename)
		if err != nil {
			return err
		}
	}

	fmt.Println("分片完成，所有 SHA256 哈希写入成功。")
	return nil
}

func readHashFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hashes := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "  ")
		if len(parts) == 2 {
			hashes[parts[1]] = parts[0]
		}
	}
	return hashes, scanner.Err()
}
