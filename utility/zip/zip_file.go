package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipFile(source string) (string, int64, error) {
	target := strings.TrimSuffix(source, filepath.Ext(source)) + ".zip"
	zipFile, err := os.Create(target)
	if err != nil {
		return target, 0, err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	file, err := os.Open(source)
	if err != nil {
		return target, 0, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return target, 0, err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return target, 0, err
	}
	header.Name = info.Name()
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return target, 0, err
	}
	_, err = io.Copy(writer, file)
	_ = zipWriter.Close()
	info, err = zipFile.Stat()
	if err != nil {
		return target, 0, err
	}
	return target, info.Size(), nil
}
