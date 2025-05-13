package facades

import (
	"github.com/goravel/framework/contracts/filesystem"

	"github.com/goravel/minio"
)

func Minio(disk string) (filesystem.Driver, error) {
	instance, err := minio.App.MakeWith(minio.Binding, map[string]any{"disk": disk})
	if err != nil {
		return nil, err
	}

	return instance.(*minio.Minio), nil
}
