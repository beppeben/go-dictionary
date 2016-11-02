package web

import (
	"strconv"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestStats(t *testing.T) {

	user := &User{Ip: "89.3.117.15", Counter: 10}

	/*
		user, err := getUserFromIp("89.3.117.15")

		if err != nil {
			log.Println(err.Error())
		}
	*/
	user.getLocation()

	log.Println(user.City + " - " + strconv.FormatInt(user.Counter, 10))

}
