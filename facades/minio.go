package facades

import (
	"log"

	"github.com/goravel/framework/contracts/filesystem"

	"github.com/goravel/minio"
)

func Minio() filesystem.Driver {
	instance, err := minio.App.Make(minio.Binding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*minio.Minio)
}
