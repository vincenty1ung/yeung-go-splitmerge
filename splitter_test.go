package main

import (
	"context"
	"fmt"
	"testing"
)

func TestSplitFileConcurrent(t *testing.T) {
	ctx := context.Background()

	err := SplitFileConcurrent(ctx, "./demo/file/20250506.zip", "./demo/chunks", 30*1024*1024)
	if err != nil {
		fmt.Println("分片失败：", err)
		return
	}

}
