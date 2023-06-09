# minio

A minio disk driver for facades.Storage of Goravel.

## Install

1. Add package

```
go get -u github.com/goravel/minio
```

2. Register service provider

```
// config/app.go
import "github.com/goravel/minio"

"providers": []foundation.ServiceProvider{
    ...
    &minio.ServiceProvider{},
}
```

3. Publish configuration file
dd
```
go run . artisan vendor:publish --package=github.com/goravel/minio
```

4. Fill your minio configuration to `config/minio.go` file

5. Add minio disk to `config/filesystems.go` file

```
// config/filesystems.go
import (
    "github.com/goravel/framework/contracts/filesystem"
    miniofacades "github.com/goravel/minio/facades"
)

"disks": map[string]any{
    ...
    "minio": map[string]any{
        "driver": "custom",
        "via": func() (filesystem.Driver, error) {
            return miniofacades.Minio(), nil
        },
    },
}
```

## Testing

Run command below to run test(fill your owner minio configuration):

```
TENCENT_ACCESS_KEY_ID= TENCENT_ACCESS_KEY_SECRET= TENCENT_BUCKET= TENCENT_URL= go test ./...
```
