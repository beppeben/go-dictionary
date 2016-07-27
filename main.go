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

	handler := persistence.NewMySqlHandler(config)
	reader := excel.NewReader(config.GetExcelDir() + "mydb.xlsx")
	repo := persistence.NewRepo(handler, reader)

	webhandler := web.NewWebHandler(repo, config, sysutils)
	webhandler.StartServer()

	select {}
}
