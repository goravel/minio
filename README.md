# Minio

A Minio disk driver for facades.Storage of Goravel.

## Version

| goravel/minio | goravel/framework |
|---------------|-------------------|
| v1.3.*        | v1.15.*           |
| v1.2.*        | v1.14.*           |
| v1.1.*        | v1.13.*           |
| v1.0.*        | v1.12.*           |

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

3. Add minio disk to `config/filesystems.go` file

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
        "key":      config.Env("MINIO_ACCESS_KEY_ID"),
        "secret":   config.Env("MINIO_ACCESS_KEY_SECRET"),
        "region":   config.Env("MINIO_REGION"),
        "bucket":   config.Env("MINIO_BUCKET"),
        "url":      config.Env("MINIO_URL"),
        "endpoint": config.Env("MINIO_ENDPOINT"),
        "ssl":      config.Env("MINIO_SSL", false),
        "via": func() (filesystem.Driver, error) {
            return miniofacades.Minio("minio"), nil // The `minio` value is the `disks` key
        },
    },
}
```

## Testing

Run command below to run test(fill your owner minio configuration):

```
MINIO_ACCESS_KEY_ID= MINIO_ACCESS_KEY_SECRET= MINIO_BUCKET= go test ./...
```
