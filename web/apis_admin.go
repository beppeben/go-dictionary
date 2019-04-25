package web

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func (handler WebserviceHandler) DeployFront(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("bundle")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "ERROR_BAD_FILE")
		return
	}
	defer file.Close()
	err = handler.sutils.ExtractZipToHttpDir(file, r.ContentLength)
	if err == nil {
		fmt.Fprintf(w, "OK")
	} else {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "ERROR")
	}
}

func (handler WebserviceHandler) DeployCal(w http.ResponseWriter, r *http.Request) {
	log.Debug("Receiving calendar file")
	file, _, err := r.FormFile("bundle")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "Error receiving excel file: %v", err)
		return
	}
	log.Debug("Copying file to folder")
	err = handler.sutils.CopyFileToExcelDir(file, "calendar.xlsx")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "Error copying excel file: %v", err)
		return
	}
	err = handler.repo.ResetCalendar()
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "Error resetting calendar: %v", err)
		return
	}
	fmt.Fprintf(w, "OK")
}

func (handler WebserviceHandler) DeployDb(w http.ResponseWriter, r *http.Request) {
	log.Debug("Receiving db file")
	file, _, err := r.FormFile("bundle")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "Error receiving excel file: %v", err)
		return
	}
	defer file.Close()
	log.Debug("Copying file to folder")
	err = handler.sutils.CopyFileToExcelDir(file, "mydb.xlsx")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "Error copying excel file: %v", err)
		return
	}
	err = handler.repo.ResetDB()
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "Error resetting database: %v", err)
		return
	}
	fmt.Fprintf(w, "OK")
}
