package utils

import (
	"mime/multipart"
)

type SysConfig interface {
	GetHTTPDir() string
	GetExcelDir() string
}

type SysUtils struct {
	config SysConfig
}

func NewSysUtils(config SysConfig) *SysUtils {
	return &SysUtils{config}
}

func (u *SysUtils) ExtractZipToHttpDir(file multipart.File, length int64) error {
	return ExtractZipToDir(file, length, u.config.GetHTTPDir())
}

func (u *SysUtils) CopyFileToExcelDir(file multipart.File) error {
	return CopyFileToPath(file, u.config.GetExcelDir(), "mydb.xlsx")
}
