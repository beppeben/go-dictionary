package excel

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestExcel(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})

	reader := NewReader("/home/giuseppe/Desktop/jewels.xlsx")
	reader.RefreshFile()
	matrix, err := reader.GetMatrix("arabic")
	if err != nil {
		log.Debug(err.Error())
	}
	log.Debugln(matrix[1][0])
	log.Debugln(matrix[2][0])
	log.Debugln(matrix[3][0])
}
