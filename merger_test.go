package main

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestMergeChunksConcurrent(t *testing.T) {
	ctx := context.Background()
	err := MergeChunksConcurrent(ctx, "./demo/chunks", "./demo/merged/merged_20250506.zip")
	if err != nil {
		fmt.Println("合并失败：", err)
	}
}

func TestCalcFileSHA256(t *testing.T) {
	originalHash := CalcFileSHA256("./demo/file/20250506.zip")
	mergedHash := CalcFileSHA256("./demo/merged/merged_20250506.zip")
	if originalHash != mergedHash {
		fmt.Println("文件合并后与原始文件不一致")
	}
	err := os.Remove("./demo/merged/merged_20250506.zip")
	if err != nil {
		fmt.Println(err)
		return
	}
}
