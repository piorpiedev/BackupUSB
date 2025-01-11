package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Tar(paths []string, out io.Writer) (files, folders uint64, err error) {
	tarWriter := tar.NewWriter(out)
	defer tarWriter.Close()
	files, folders = 0, 0

	for _, fpath := range paths {
		fpath = filepath.Clean(fpath)

		info, err := os.Stat(fpath)
		if err != nil {
			return files, folders, err // File does not exist or is inaccessible
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(fpath)
		}

		err = filepath.Walk(fpath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				fileHeader, err := tar.FileInfoHeader(info, info.Name())
				if err != nil {
					return err
				}

				if baseDir != "" {
					fileHeader.Name = filepath.Join(baseDir, strings.TrimPrefix(path, fpath))
				}

				if err := tarWriter.WriteHeader(fileHeader); err != nil {
					return err
				}

				if info.IsDir() {
					folders++
					fmt.Println("+ " + fileHeader.Name)
					return nil
				}

				fmt.Printf("+ [%s] %s\n", FormatByteCount(info.Size()), fileHeader.Name)
				files++
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				if _, err = io.Copy(tarWriter, file); err != nil {
					return err
				}
				return nil
			})

		if err != nil {
			return files, folders, err
		}
	}
	return files, folders, nil
}

func Untar(in io.Reader, out string) (files, folders uint64, err error) {
	tarReader := tar.NewReader(in)
	files, folders = 0, 0

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return files, folders, err
		}

		path := filepath.Join(out, header.Name)
		info := header.FileInfo()

		if info.IsDir() {
			fmt.Println("+ " + header.Name)
			folders++
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return files, folders, err
			}
			continue
		}

		fmt.Printf("+ [%s] %s\n", FormatByteCount(info.Size()), header.Name)
		files++
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return files, folders, err
		}
		defer file.Close()
		
		if _, err = io.Copy(file, tarReader); err != nil {
			return files, folders, err
		}
	}
	return files, folders, nil
}
