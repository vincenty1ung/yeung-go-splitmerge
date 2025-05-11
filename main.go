package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cast"
)

// main.go 提供示例命令行调用，根据参数执行拆分或合并操作
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: splitmerge <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  split <inputfile> <outputDir> <chunkSize>    Split a file default into 1MB parts")
		fmt.Println("  merge [hashfile]     Merge files from hashes.sha256 (default hashes.sha256)")
		os.Exit(1)
	}
	cmd := os.Args[1]
	ctx := context.Background()
	switch cmd {
	case "split":
		if len(os.Args) != 5 {
			fmt.Println(os.Args)
			fmt.Println("Usage: splitmerge split <inputfile> <outputDir> <chunkSize>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		outputDir := os.Args[3]
		chunkSize := os.Args[4]
		if err := SplitFileConcurrent(ctx, filePath, outputDir, cast.ToInt64(chunkSize)); err != nil {
			log.Fatalf("Split failed: %v", err)
		}
	case "merge":
		// var hashFile string = "hashes.sha256"
		if len(os.Args) != 4 {
			fmt.Println("Usage: splitmerge merge [hashFileDir] <outputDir>")
			os.Exit(1)
		}
		hashFileDir := os.Args[2]
		outputDir := os.Args[3]
		if err := MergeChunksConcurrent(ctx, hashFileDir, outputDir); err != nil {
			log.Fatalf("Merge failed: %v", err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
