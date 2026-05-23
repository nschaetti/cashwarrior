package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cash <command>")
	}

	cmd := os.Args[1]

	switch cmd {
	case "add":
		amount := os.Args[2]
		currency := os.Args[3]
		product := os.Args[4]
		var shop string
		if len(os.Args) > 5 {
			shop = os.Args[5]
		}
		fmt.Println("Adding money")
	default:
		fmt.Println("Unknown command")
	}
}
