package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	uuid "github.com/satori/go.uuid"
)

func GenerateUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func IsValidIPAddress(addr string) bool {
	return net.ParseIP(addr) != nil
}

func IsValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
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
