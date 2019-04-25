package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/beppeben/go-dictionary/excel"
	"github.com/beppeben/go-dictionary/persistence"
	"github.com/beppeben/go-dictionary/utils"
	"github.com/beppeben/go-dictionary/web"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})

	config := utils.NewAppConfig()
	sysutils := utils.NewSysUtils(config)
	emailutils := utils.NewEmailUtils(config)

	handler := persistence.NewMySqlHandler(config)
	dbReader := excel.NewReader(config.GetExcelDir() + "mydb.xlsx")
	calReader := excel.NewReader(config.GetExcelDir() + "calendar.xlsx")
	repo := persistence.NewRepo(handler, dbReader, calReader)

	webhandler := web.NewWebHandler(repo, config, sysutils, emailutils)
	webhandler.StartServer()

	select {}
}
