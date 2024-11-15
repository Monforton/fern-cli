package main

import (
	"fmt"
	"os"

	"github.com/monforton/fern-cli/cmd/fern"
)

func main() {
	// Do Stuff Here
	banner, err := os.ReadFile("./static/banner.txt")
	if err == nil {
		fmt.Println(string(banner))
	}
	fern.Execute()
}
