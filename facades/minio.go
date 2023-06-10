package facades

import (
	"log"

	"github.com/goravel/framework/contracts/filesystem"

	"github.com/goravel/minio"
)

func Minio(disk string) filesystem.Driver {
	instance, err := minio.App.MakeWith(minio.Binding, map[string]any{"disk": disk})
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*minio.Minio)
}
