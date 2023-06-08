package minio

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/support/file"
)

func fullPathOfFile(filePath string, source filesystem.File, name string) (string, error) {
	extension := path.Ext(name)
	if extension == "" {
		var err error
		extension, err = file.Extension(source.File(), true)
		if err != nil {
			return "", err
		}

		return filepath.Join(filePath, strings.TrimSuffix(strings.TrimPrefix(path.Base(name), string(filepath.Separator)), string(filepath.Separator))+"."+extension), nil
	} else {
		return filepath.Join(filePath, strings.TrimPrefix(path.Base(name), string(filepath.Separator))), nil
	}
}

func validPath(path string) string {
	realPath := strings.TrimPrefix(path, "."+string(filepath.Separator))
	realPath = strings.TrimPrefix(realPath, string(filepath.Separator))
	realPath = strings.TrimPrefix(realPath, ".")
	if realPath != "" && !strings.HasSuffix(realPath, string(filepath.Separator)) {
		realPath += string(filepath.Separator)
	}

	return realPath
}
