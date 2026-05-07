package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Nykenik24/oxy/lexer"
)

func main() {
	start := time.Now()

	l := lexer.New()
	input := `123 456.789 hello _world "string 123 hellooo" 'c' if`

	tokens, err := l.Lex(input)
	if err != nil {
		log.Fatal(err)
	}

	for _, tk := range tokens {
		fmt.Println(tk)
	}

	elapsed := time.Since(start)
	fmt.Println()
	fmt.Printf("Took \x1b[34m%v\x1b[0m to lex\n", elapsed)
}
