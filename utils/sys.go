package utils

import (
	"mime/multipart"
)

type SysConfig interface {
	GetHTTPDir() string
	GetExcelDir() string
}

type Sys struct {
	config SysConfig
}

func NewSysUtils(config SysConfig) *Sys {
	return &Sys{config}
}

func (u *Sys) ExtractZipToHttpDir(file multipart.File, length int64) error {
	return ExtractZipToDir(file, length, u.config.GetHTTPDir())
}

func (u *Sys) CopyFileToExcelDir(file multipart.File) error {
	return CopyFileToPath(file, u.config.GetExcelDir(), "mydb.xlsx")
}
