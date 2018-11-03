package checker

import (
	"crypto/md5"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Checker interface {
	List()
	Get()
	Compare()
}

type FileChecker struct {
}

type HashRes struct {
	FileName string
	Hash     string
}

type FileCheckerPath struct {
	Path  string
	Skips map[string]bool
}

// List directories and files with their hashes.
func (fc FileChecker) List(root string, skips map[string]bool) (hr []HashRes, err error) {
	// TODO mora postojati skip dir lista
	err = filepath.Walk(root, func(path string, info os.FileInfo, err0 error) error {
		if info.IsDir() {
			if skips[path] {
				return filepath.SkipDir
			}
			hr = append(hr, HashRes{FileName: path})
		} else {
			f, err2 := os.Open(path)
			if err2 != nil {
				return err2
			}
			defer f.Close()
			h := md5.New()
			if _, err3 := io.Copy(h, f); err3 != nil {
				log.Fatal(err3)
			}
			hr = append(hr, HashRes{FileName: path, Hash: string(h.Sum(nil))})
		}
		return nil
	})
	return hr, err
}
