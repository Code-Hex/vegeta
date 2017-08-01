package main

import (
	"os"

	"github.com/Code-Hex/vegeta"
)

func main() {
	os.Exit(vegeta.New().Run())
}
