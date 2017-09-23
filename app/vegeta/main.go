package main

import (
	"os"

	"github.com/Code-Hex/vegeta"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	os.Exit(vegeta.New().Run())
}
