package main

import (
	"github.com/Nykenik24/oxy/internal/runner"
)

func main() {
	run := runner.New()

	run.Benchmark(func() {
		run.RunString(`module main
variant StringOrInt {
	string s
	int i
}`)
	})
}
