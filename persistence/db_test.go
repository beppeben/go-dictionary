package persistence

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/beppeben/go-dictionary/utils"
)

func TestDb(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	config := utils.NewAppConfig()
	handler := NewMySqlHandler(config)
	db := handler.Connection

}
