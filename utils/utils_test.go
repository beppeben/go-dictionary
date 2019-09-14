package utils

import (
	//"fmt"
	"testing"
)

func TestStrings(t *testing.T) {
	config := NewConfig("../config/")
	msgutils := NewMessageUtils(config)
	msgutils.SendToSlack("test")
}
