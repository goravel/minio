package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	config := `map[string]any{
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

	appConfigPath := path.Config("app.go")
	filesystemsConfigPath := path.Config("filesystems.go")
	moduleImport := setup.Paths().Module().Import()
	minioServiceProvider := "&minio.ServiceProvider{}"
	filesystemContract := "github.com/goravel/framework/contracts/filesystem"
	minioFacades := "github.com/goravel/minio/facades"
	filesystemsDisksConfig := match.Config("filesystems.disks")
	filesystemsConfig := match.Config("filesystems")

	setup.Install(
		// Add minio service provider to app.go if not using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Imports()).Modify(modify.AddImport(moduleImport)).
			Find(match.Providers()).Modify(modify.Register(minioServiceProvider))),

		// Add minio service provider to providers.go if using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.AddProviderApply(moduleImport, minioServiceProvider)),

		// Add minio disk to filesystems.go
		modify.GoFile(filesystemsConfigPath).Find(match.Imports()).Modify(
			modify.AddImport(filesystemContract),
			modify.AddImport(minioFacades, "miniofacades"),
		).
			Find(filesystemsDisksConfig).Modify(modify.AddConfig("minio", config)).
			Find(filesystemsConfig).Modify(modify.AddConfig("default", `"minio"`)),
	).Uninstall(
		// Remove minio disk from filesystems.go
		modify.GoFile(filesystemsConfigPath).
			Find(filesystemsConfig).Modify(modify.AddConfig("default", `"local"`)).
			Find(filesystemsDisksConfig).Modify(modify.RemoveConfig("minio")).
			Find(match.Imports()).Modify(
			modify.RemoveImport(filesystemContract),
			modify.RemoveImport(minioFacades, "miniofacades"),
		),

		// Remove minio service provider from app.go if not using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Providers()).Modify(modify.Unregister(minioServiceProvider)).
			Find(match.Imports()).Modify(modify.RemoveImport(moduleImport))),

		// Remove minio service provider from providers.go if using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.RemoveProviderApply(moduleImport, minioServiceProvider)),
	).Execute()
}
