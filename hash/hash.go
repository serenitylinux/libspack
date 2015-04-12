package hash

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func Md5sum(filename string) (string, error) {
	h := md5.New()
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type HashList map[string]string
