package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

var config = `map[string]any{
        "driver": "custom",
        "key":      config.Env("MINIO_ACCESS_KEY_ID"),
        "secret":   config.Env("MINIO_ACCESS_KEY_SECRET"),
        "region":   config.Env("MINIO_REGION"),
        "bucket":   config.Env("MINIO_BUCKET"),
        "url":      config.Env("MINIO_URL"),
        "endpoint": config.Env("MINIO_ENDPOINT"),
        "ssl":      config.Env("MINIO_SSL", false),
        "via": func() (filesystem.Driver, error) {
            return miniofacades.Minio("minio") // The ` + "`minio`" + ` value is the ` + "`disks`" + ` key
        },
    }`

func main() {
	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&minio.ServiceProvider{}")),
			modify.GoFile(path.Config("filesystems.go")).
				Find(match.Imports()).Modify(modify.AddImport("github.com/goravel/framework/contracts/filesystem"), modify.AddImport("github.com/goravel/minio/facades", "miniofacades")).
				Find(match.Config("filesystems.disks")).Modify(modify.AddConfig("minio", config)),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Unregister("&minio.ServiceProvider{}")),
			modify.GoFile(path.Config("filesystems.go")).
				Find(match.Config("filesystems.disks")).Modify(modify.RemoveConfig("minio")).
				Find(match.Imports()).Modify(modify.RemoveImport("github.com/goravel/framework/contracts/filesystem"), modify.RemoveImport("github.com/goravel/minio/facades", "miniofacades")),
		).
		Execute()
}
