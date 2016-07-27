package excel

import (
	"errors"
	"strings"

	"github.com/tealeg/xlsx"
)

type ExcelReader struct {
	xlFilePath string
	xlFile     *xlsx.File
}

func NewReader(xlFilePath string) *ExcelReader {
	if xlFilePath == "" {
		panic("Bad file name: " + xlFilePath)
	}
	reader := &ExcelReader{}
	reader.xlFilePath = xlFilePath
	return reader
}

func (e *ExcelReader) RefreshFile() error {
	xlFile, err := xlsx.OpenFile(e.xlFilePath)
	if err != nil {
		return errors.New("Invalid file: " + err.Error())
	}
	e.xlFile = xlFile
	return nil
}

func (e *ExcelReader) GetSheet(name string) (*xlsx.Sheet, error) {
	for _, sheet := range e.xlFile.Sheets {
		if sheet.Name == name {
			return sheet, nil
		}
	}
	return nil, errors.New("No worksheet named " + name)
}

func (e *ExcelReader) GetMatrix(title string) ([][]string, error) {
	sheet, err := e.GetSheet(title)
	if err != nil {
		return nil, err
	}
	cols := 0
	for _, cell := range sheet.Rows[0].Cells {
		value, err := cell.String()
		if err != nil || value == "" {
			break
		}
		cols++
	}
	result := make([][]string, 0)
	for _, row := range sheet.Rows {
		temp := make([]string, cols)
		allEmpty := true
		for j := 0; j < cols && j < len(row.Cells); j++ {
			value, err := row.Cells[j].String()
			if err != nil {
				return nil, err
			}
			temp[j] = strings.TrimSpace(value)
			allEmpty = allEmpty && (temp[j] == "")
		}
		if allEmpty {
			break
		}
		result = append(result, temp)
	}
	return result, nil
}
