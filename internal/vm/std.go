package vm

import "fmt"

var std = StandardLibrary{
	"println": func(args []any) []any {
		for _, arg := range args {
			fmt.Print(arg)
			fmt.Print("\t")
		}
		fmt.Println()
		return nil
	},
}

func GetSTD() StandardLibrary {
	return std
}
