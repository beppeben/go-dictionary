package utils

import (
	"fmt"
	"testing"
)

func TestStrings(t *testing.T) {
	fmt.Println(MapToASCII("Clément"))
	fmt.Println(MapToASCII("àéùciaoç"))
}
