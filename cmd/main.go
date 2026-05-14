package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Nykenik24/oxy/internal/vm"
)

func puttime(elapsed time.Duration) {
	fmt.Printf("took \x1b[34m~%s\x1b[0m \x1b[90m(%s)\x1b[0m to run\n", elapsed.Round(time.Microsecond), elapsed)
}

func benchmark(fn func()) {
	start := time.Now()

	fn()

	elapsed := time.Since(start)
	puttime(elapsed)
}

func multibench(fn func(), iter int) {
	start := time.Now()

	for range iter {
		fn()
	}

	elapsed := time.Since(start)
	puttime(elapsed)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: oxy <filename>")
		os.Exit(1)
	}

	file := os.Args[1]
	v := vm.New()
	v.LoadFile(file)
	// multibench(func() {
	// 	v.Run()
	// }, 100)

	v.SetDebug(vm.DebugAST)
	v.Run()
}
