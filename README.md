# Minio

A Minio disk driver for facades.Storage of Goravel.

## Install

Run the command below in your project to install the package automatically:

```
./artisan package:install github.com/goravel/minio
```

Or check [the setup file](./setup/setup.go) to install the package manually.

## Testing

Run command below to run test:

```
MINIO_ACCESS_KEY_ID= MINIO_ACCESS_KEY_SECRET= MINIO_BUCKET= go test ./...
```
