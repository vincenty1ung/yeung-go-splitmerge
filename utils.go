package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const PartSize int64 = 1 << 20 // 1MB

// hashToString 计算 data 的 SHA256 哈希并返回十六进制字符串
func hashToString(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

// min 返回 a,b 中的最小值
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func CalcFileSHA256(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "ERROR"
	}
	defer file.Close()

	hasher := sha256.New()
	io.Copy(hasher, file)
	return hex.EncodeToString(hasher.Sum(nil))
}
