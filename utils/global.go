package utils

import (
	"archive/zip"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	//log "github.com/Sirupsen/logrus"
)

func CopyFileToPath(file multipart.File, dir string, filename string) error {
	os.MkdirAll(dir, 0755)
	os.Remove(dir + filename)
	f, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	return err
}

func Contains(slice []string, term string) bool {
	for _, s := range slice {
		if s == term {
			return true
		}
	}
	return false
}

func ExtractZipToDir(file multipart.File, length int64, dest string) error {
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
