package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipFiles(sources []string) (string, int64, error) {
	if len(sources) == 0 {
		return "", 0, fmt.Errorf("no files to zip")
	}
	target := strings.TrimSuffix(sources[0], filepath.Ext(sources[0])) + ".zip"
	zipFile, err := os.Create(target)
	if err != nil {
		return target, 0, err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	for _, source := range sources {
		file, err := os.Open(source)
		if err != nil {
			zipWriter.Close()
			return target, 0, err
		}
		info, err := file.Stat()
		if err != nil {
			file.Close()
			zipWriter.Close()
			return target, 0, err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			file.Close()
			zipWriter.Close()
			return target, 0, err
		}
		header.Name = info.Name()
		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			file.Close()
			zipWriter.Close()
			return target, 0, err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			zipWriter.Close()
			return target, 0, err
		}
	}
	_ = zipWriter.Close()
	info, err := zipFile.Stat()
	if err != nil {
		return target, 0, err
	}
	return target, info.Size(), nil
}
