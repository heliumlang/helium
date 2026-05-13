package main

import (
	"fmt"
	"os"

	"github.com/Nykenik24/oxy/internal/runner"
)

func main() {
	run := runner.New()

	if len(os.Args) < 2 {
		fmt.Println("Usage: oxy <filename>")
		os.Exit(1)
	}

	file := os.Args[1]

	run.Benchmark(func() {
		if err := run.RunFile(file); err != nil {
			err.SetFilename(file).Print()
		}
	})
}
