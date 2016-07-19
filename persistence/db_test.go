package persistence

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/beppeben/go-dictionary/excel"
	"github.com/beppeben/go-dictionary/utils"
)

func TestDb(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	config := utils.NewAppConfig()
	handler := NewMySqlHandler(config)
	reader := excel.NewReader("/home/giuseppe/Desktop/jewels.xlsx")
	repo := NewRepo(handler, reader)
	/*
		err := repo.ResetDB()
		if err != nil {
			log.Debug(err.Error())
		}

			words, err := repo.GetAllWords("french", "arabic")
			if err != nil {
				log.Debug(err.Error())
			} else {
				for _, word := range words {
					log.Debug(word.Word + "-" + word.LangTag)
				}
			}
	*/
	words, err := repo.Search("col de cygne", "french", "italian")
	if err != nil {
		log.Debug(err.Error())
	}
	for _, w := range words {
		for _, t := range w.Translations {
			log.Debug(t.Word)
		}
	}
}
