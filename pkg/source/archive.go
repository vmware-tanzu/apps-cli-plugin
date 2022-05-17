/*
Copyright 2022 VMware, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func HandleZip(fileName string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	defer file.Close()
	if err = ExtractZip(file, info.Size(), tmpDir); err != nil {
		return "", err
	}
	return tmpDir, nil
}

func ExtractZip(reader io.ReaderAt, size int64, dir string) error {
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		filePath := filepath.Join(dir, file.Name)
		fileMode := file.Mode()
		if isFatFile(file.FileHeader) {
			fileMode = 0777
		}

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, fileMode)
			if err != nil {
				return err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, srcFile); err != nil {
			return err
		}

		if err := outFile.Close(); err != nil {
			return err
		}

		if err := srcFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

func IsZip(fileName string) bool {
	file, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer file.Close()

	// http://golang.org/pkg/net/http/#DetectContentType
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return false
	}

	return http.DetectContentType(buf) == "application/zip"
}

func isFatFile(header zip.FileHeader) bool {
	var (
		creatorFAT  uint16 = 0
		creatorVFAT uint16 = 14
	)

	// This identifies FAT files, based on the `zip` source: https://golang.org/src/archive/zip/struct.go
	firstByte := header.CreatorVersion >> 8
	return firstByte == creatorFAT || firstByte == creatorVFAT
}

func IsDir(fileName string) bool {
	file, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.IsDir()
}
