package utils

import (
	"archive/zip"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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
	return extractZipToDir(file, length, u.config.GetHTTPDir())
}

func (u *Sys) CopyFileToExcelDir(file multipart.File) error {
	return copyFileToPath(file, u.config.GetExcelDir(), "mydb.xlsx")
}

func copyFileToPath(file multipart.File, dir string, filename string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	err = os.Remove(dir + filename)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	return err
}

func extractZipToDir(file multipart.File, length int64, dest string) error {
	r, err := zip.NewReader(file, length)
	if err != nil {
		return err
	}
	os.MkdirAll(dest, 0755)

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
