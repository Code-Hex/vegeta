package utils

import (
	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"
)

func GenerateUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
