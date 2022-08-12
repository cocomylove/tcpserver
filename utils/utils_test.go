package utils

import (
	"fmt"
	"testing"
)

func TestGetConnectionID(t *testing.T) {
	for i := 0; i < 101; i++ {
		id := GetConnectionID()
		fmt.Println(id)
	}

}
